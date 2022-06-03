package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/labd/contentstack-go-sdk/management"
)

type resourceLocaleType struct{}

type LocaleData struct {
	UID            types.String `tfsdk:"uid"`
	Name           types.String `tfsdk:"name"`
	Code           types.String `tfsdk:"code"`
	FallbackLocale types.String `tfsdk:"fallback_locale"`
}

// Global Field Resource schema
func (r resourceLocaleType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
		Contentstack offers multilingual support, which allows you to create
		entries in any language of your choice. When creating entries in other
		languages, they inherit data initially from the fallback language until
		they are localized.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"uid": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"code": {
				Type:     types.StringType,
				Optional: true,
			},
			"fallback_locale": {
				Type:     types.StringType,
				Optional: true,
			},
		},
	}, nil
}

// New resource instance
func (r resourceLocaleType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceLocale{
		p: *(p.(*provider)),
	}, nil
}

type resourceLocale struct {
	p provider
}

func (r resourceLocale) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan LocaleData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewLocaleInput(&plan)
	resource, err := r.p.stack.LocaleCreate(ctx, *input)
	if err != nil {
		diags := processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(resource, input)
	resp.Diagnostics.Append(diags...)

	// Write to state.
	state := NewLocaleData(resource)
	diags = MergeLocaleResponse(state, &plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceLocale) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state LocaleData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.p.stack.LocaleFetch(ctx, state.Code.Value)
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving locale",
				fmt.Sprintf("The locale %s was not found.", state.Code.Value))
			resp.Diagnostics.Append(d)
		} else {
			diags := processRemoteError(err)
			resp.Diagnostics.Append(diags...)
		}
		return
	}

	curr := NewLocaleInput(&state)
	diags = processResponse(resource, curr)
	resp.Diagnostics.Append(diags...)

	// Set state
	newState := NewLocaleData(resource)
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r resourceLocale) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state LocaleData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.LocaleDelete(ctx, state.Code.Value)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceLocale) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan LocaleData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state LocaleData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewLocaleInput(&plan)
	resource, err := r.p.stack.LocaleUpdate(ctx, state.Code.Value, *input)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(resource, input)
	resp.Diagnostics.Append(diags...)

	// Set state
	result := NewLocaleData(resource)
	diags = MergeLocaleResponse(result, &plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r resourceLocale) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func NewLocaleData(field *management.Locale) *LocaleData {
	state := &LocaleData{
		UID:            types.String{Value: field.UID},
		Name:           types.String{Value: field.Name},
		Code:           types.String{Value: field.Code},
		FallbackLocale: types.String{Value: field.FallbackLocale},
	}
	return state
}

func NewLocaleInput(field *LocaleData) *management.LocaleInput {

	input := &management.LocaleInput{
		Name:           field.Name.Value,
		Code:           field.Code.Value,
		FallbackLocale: field.FallbackLocale.Value,
	}

	return input
}

func MergeLocaleResponse(out *LocaleData, in *LocaleData) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.FallbackLocale != out.FallbackLocale {
		diags.AddAttributeWarning(
			tftypes.NewAttributePath().WithAttributeName("fallback_locale"),
			"Contentstack modified fallback_locale",
			fmt.Sprintf(
				"Contentstack set the fallback_locale to a different value then requested. Requested was %s but value is %s",
				in.FallbackLocale.Value, out.FallbackLocale.Value))
	}
	return diags
}
