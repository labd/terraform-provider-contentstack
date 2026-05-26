package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/contentstack-go-sdk/management"
)

type LocaleData struct {
	UID            types.String `tfsdk:"uid"`
	Name           types.String `tfsdk:"name"`
	Code           types.String `tfsdk:"code"`
	FallbackLocale types.String `tfsdk:"fallback_locale"`
}

type resourceLocale struct {
	p *contentstackProvider
}

// Metadata
func (r *resourceLocale) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "contentstack_locale"
}

// Global Field Resource schema
func (r *resourceLocale) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
		Contentstack offers multilingual support, which allows you to create
		entries in any language of your choice. When creating entries in other
		languages, they inherit data initially from the fallback language until
		they are localized.
		`,
		Attributes: map[string]schema.Attribute{
			"uid": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"code": schema.StringAttribute{
				Optional: true,
			},
			"fallback_locale": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *resourceLocale) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

func (r *resourceLocale) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LocaleData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.p.stack.LocaleFetch(ctx, state.Code.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving locale",
				fmt.Sprintf("The locale %s was not found.", state.Code.ValueString()))
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

func (r *resourceLocale) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LocaleData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.LocaleDelete(ctx, state.Code.ValueString())
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *resourceLocale) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	resource, err := r.p.stack.LocaleUpdate(ctx, state.Code.ValueString(), *input)
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

func (r *resourceLocale) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uid"), req, resp)
}

func NewLocaleData(field *management.Locale) *LocaleData {
	state := &LocaleData{
		UID:            types.StringValue(field.UID),
		Name:           types.StringValue(field.Name),
		Code:           types.StringValue(field.Code),
		FallbackLocale: types.StringValue(field.FallbackLocale),
	}
	return state
}

func NewLocaleInput(field *LocaleData) *management.LocaleInput {

	input := &management.LocaleInput{
		Name:           field.Name.ValueString(),
		Code:           field.Code.ValueString(),
		FallbackLocale: field.FallbackLocale.ValueString(),
	}

	return input
}

func MergeLocaleResponse(out *LocaleData, in *LocaleData) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.FallbackLocale != out.FallbackLocale {
		diags.AddAttributeWarning(
			path.Root("fallback_locale"),
			"Contentstack modified fallback_locale",
			fmt.Sprintf(
				"Contentstack set the fallback_locale to a different value then requested. Requested was %s but value is %s",
				in.FallbackLocale.ValueString(), out.FallbackLocale.ValueString()))
	}
	return diags
}
