package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/labd/contentstack-go-sdk/management"
)

type resourceContentTypeType struct{}

type ContentTypeData struct {
	UID         types.String `tfsdk:"uid"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Schema      types.String `tfsdk:"schema"`
}

// Global Field Resource schema
func (r resourceContentTypeType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
		Content type defines the structure or schema of a page or a section of
		your web or mobile property. To create content for your application, you
		are required to first create a content type, and then create entries
		using the content type.

		Note: Removing a field or modifying its properties may result in data
		loss or invalidate field visibility rules.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"uid": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"title": {
				Type:     types.StringType,
				Required: true,
			},
			"description": {
				Type:     types.StringType,
				Optional: true,
			},
			"schema": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The schema as JSON. Use jsonencode(jsonecode(<schema>)) to work around wrong changes.",
			},
		},
	}, nil
}

// New resource instance
func (r resourceContentTypeType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceContentType{
		p: *(p.(*provider)),
	}, nil
}

type resourceContentType struct {
	p provider
}

func (r resourceContentType) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan ContentTypeData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewContentTypeInput(&plan)
	resource, err := r.p.stack.ContentTypeCreate(ctx, *input)
	if err != nil {
		diags := processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(resource, input)
	resp.Diagnostics.Append(diags...)

	// Write to state.
	state := NewContentTypeData(resource)
	MergeContentType(state, &plan)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceContentType) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state ContentTypeData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.p.stack.ContentTypeFetch(ctx, state.UID.Value)
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving global field",
				fmt.Sprintf("The global field with UID %s was not found.", state.UID.Value))
			resp.Diagnostics.Append(d)
		} else {
			diags := processRemoteError(err)
			resp.Diagnostics.Append(diags...)
		}
		return
	}

	curr := NewContentTypeInput(&state)
	diags = processResponse(resource, curr)
	resp.Diagnostics.Append(diags...)

	// Set state
	newState := NewContentTypeData(resource)
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r resourceContentType) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ContentTypeData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.ContentTypeDelete(ctx, state.UID.Value)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceContentType) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan ContentTypeData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state ContentTypeData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewContentTypeInput(&plan)
	resource, err := r.p.stack.ContentTypeUpdate(ctx, state.UID.Value, *input)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(resource, input)
	resp.Diagnostics.Append(diags...)

	// Set state
	result := NewContentTypeData(resource)
	MergeContentType(result, &plan)
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r resourceContentType) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func NewContentTypeData(field *management.ContentType) *ContentTypeData {

	schemaContent, err := field.Schema.MarshalJSON()
	if err != nil {
		panic(err)
	}

	state := &ContentTypeData{
		UID:         types.String{Value: field.UID},
		Title:       types.String{Value: field.Title},
		Description: types.String{Value: field.Description},
		Schema:      types.String{Value: string(schemaContent)},
	}
	return state
}

func NewContentTypeInput(field *ContentTypeData) *management.ContentTypeInput {

	input := &management.ContentTypeInput{
		UID:         &field.UID.Value,
		Title:       &field.Title.Value,
		Description: &field.Description.Value,
		Schema:      json.RawMessage(field.Schema.Value),
	}

	return input
}

func MergeContentType(out *ContentTypeData, in *ContentTypeData) {
	out.Schema = in.Schema
}
