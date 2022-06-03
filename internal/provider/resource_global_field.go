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

type resourceGlobalFieldType struct{}

type GlobalFieldData struct {
	UID               types.String `tfsdk:"uid"`
	Title             types.String `tfsdk:"title"`
	Description       types.String `tfsdk:"description"`
	MaintainRevisions types.Bool   `tfsdk:"maintain_revisions"`
	Schema            types.String `tfsdk:"schema"`
}

// Global Field Resource schema
func (r resourceGlobalFieldType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
		A Global field is a reusable field (or group of fields) that you can
		define once and reuse in any content type within your stack. This
		eliminates the need (and thereby time and efforts) to create the same
		set of fields repeatedly in multiple content types.
		`,
		Attributes: map[string]tfsdk.Attribute{
			"uid": {
				Type:     types.StringType,
				Required: true,
			},
			"title": {
				Type:     types.StringType,
				Required: true,
			},
			"maintain_revisions": {
				Type:     types.BoolType,
				Optional: true,
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
func (r resourceGlobalFieldType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceGlobalField{
		p: *(p.(*provider)),
	}, nil
}

type resourceGlobalField struct {
	p provider
}

func (r resourceGlobalField) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan GlobalFieldData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewGlobalFieldInput(&plan)
	resource, err := r.p.stack.GlobalFieldCreate(ctx, *input)
	if err != nil {
		diags := processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(resource, input)
	resp.Diagnostics.Append(diags...)

	// Write to state.
	state := NewGlobalFieldData(resource)
	MergeGlobalField(state, &plan)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceGlobalField) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state GlobalFieldData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.p.stack.GlobalFieldFetch(ctx, state.UID.Value)
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

	curr := NewGlobalFieldInput(&state)
	diags = processResponse(resource, curr)
	resp.Diagnostics.Append(diags...)

	// Set state
	newState := NewGlobalFieldData(resource)
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r resourceGlobalField) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state GlobalFieldData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.GlobalFieldDelete(ctx, state.UID.Value)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceGlobalField) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan GlobalFieldData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state GlobalFieldData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewGlobalFieldInput(&plan)
	resource, err := r.p.stack.GlobalFieldUpdate(ctx, state.UID.Value, *input)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(resource, input)
	resp.Diagnostics.Append(diags...)

	// Set state
	result := NewGlobalFieldData(resource)
	MergeGlobalField(result, &plan)
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r resourceGlobalField) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func NewGlobalFieldData(field *management.GlobalField) *GlobalFieldData {

	schemaContent, err := field.Schema.MarshalJSON()
	if err != nil {
		panic(err)
	}

	state := &GlobalFieldData{
		UID:               types.String{Value: field.UID},
		Title:             types.String{Value: field.Title},
		Description:       types.String{Value: field.Description},
		MaintainRevisions: types.Bool{Value: field.MaintainRevisions},
		Schema:            types.String{Value: string(schemaContent)},
	}
	return state
}

func NewGlobalFieldInput(field *GlobalFieldData) *management.GlobalFieldInput {

	input := &management.GlobalFieldInput{
		UID:               &field.UID.Value,
		Title:             &field.Title.Value,
		Description:       &field.Description.Value,
		MaintainRevisions: field.MaintainRevisions.Value,
		Schema:            json.RawMessage(field.Schema.Value),
	}

	return input
}

func MergeGlobalField(out *GlobalFieldData, in *GlobalFieldData) {
	out.Schema = in.Schema
}
