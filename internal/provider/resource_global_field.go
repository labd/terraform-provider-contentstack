package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/contentstack-go-sdk/management"
)

type GlobalFieldData struct {
	UID               types.String `tfsdk:"uid"`
	Title             types.String `tfsdk:"title"`
	Description       types.String `tfsdk:"description"`
	MaintainRevisions types.Bool   `tfsdk:"maintain_revisions"`
	Schema            types.String `tfsdk:"schema"`
}

type resourceGlobalField struct {
	p *contentstackProvider
}

// Metadata
func (r *resourceGlobalField) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "contentstack_global_field"
}

// Global Field Resource schema
func (r *resourceGlobalField) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
		A Global field is a reusable field (or group of fields) that you can
		define once and reuse in any content type within your stack. This
		eliminates the need (and thereby time and efforts) to create the same
		set of fields repeatedly in multiple content types.
		`,
		Attributes: map[string]schema.Attribute{
			"uid": schema.StringAttribute{
				Required: true,
			},
			"title": schema.StringAttribute{
				Required: true,
			},
			"maintain_revisions": schema.BoolAttribute{
				Optional: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"schema": schema.StringAttribute{
				Optional:    true,
				Description: "The schema as JSON. Use jsonencode(jsonecode(<schema>)) to work around wrong changes.",
			},
		},
	}
}

func (r *resourceGlobalField) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

func (r *resourceGlobalField) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GlobalFieldData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.p.stack.GlobalFieldFetch(ctx, state.UID.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving global field",
				fmt.Sprintf("The global field with UID %s was not found.", state.UID.ValueString()))
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

func (r *resourceGlobalField) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GlobalFieldData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.GlobalFieldDelete(ctx, state.UID.ValueString())
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *resourceGlobalField) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	resource, err := r.p.stack.GlobalFieldUpdate(ctx, state.UID.ValueString(), *input)
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

func (r *resourceGlobalField) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uid"), req, resp)
}

func NewGlobalFieldData(field *management.GlobalField) *GlobalFieldData {

	schemaContent, err := field.Schema.MarshalJSON()
	if err != nil {
		panic(err)
	}

	state := &GlobalFieldData{
		UID:               types.StringValue(field.UID),
		Title:             types.StringValue(field.Title),
		Description:       types.StringValue(field.Description),
		MaintainRevisions: types.BoolValue(field.MaintainRevisions),
		Schema:            types.StringValue(string(schemaContent)),
	}
	return state
}

func NewGlobalFieldInput(field *GlobalFieldData) *management.GlobalFieldInput {

	uid := field.UID.ValueString()
	title := field.Title.ValueString()
	description := field.Description.ValueString()

	input := &management.GlobalFieldInput{
		UID:               &uid,
		Title:             &title,
		Description:       &description,
		MaintainRevisions: field.MaintainRevisions.ValueBool(),
		Schema:            json.RawMessage(field.Schema.ValueString()),
	}

	return input
}

func MergeGlobalField(out *GlobalFieldData, in *GlobalFieldData) {
	out.Schema = in.Schema
}
