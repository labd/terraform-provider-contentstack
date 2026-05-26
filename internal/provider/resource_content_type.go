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

type ContentTypeData struct {
	UID         types.String `tfsdk:"uid"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	Schema      types.String `tfsdk:"schema"`
}

type resourceContentType struct {
	p *contentstackProvider
}

// Metadata
func (r *resourceContentType) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "contentstack_content_type"
}

// Content Type Resource schema
func (r *resourceContentType) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
		Content type defines the structure or schema of a page or a section of
		your web or mobile property. To create content for your application, you
		are required to first create a content type, and then create entries
		using the content type.

		Note: Removing a field or modifying its properties may result in data
		loss or invalidate field visibility rules.
		`,
		Attributes: map[string]schema.Attribute{
			"uid": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"title": schema.StringAttribute{
				Required: true,
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

func (r *resourceContentType) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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

func (r *resourceContentType) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContentTypeData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource, err := r.p.stack.ContentTypeFetch(ctx, state.UID.ValueString())
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

	curr := NewContentTypeInput(&state)
	diags = processResponse(resource, curr)
	resp.Diagnostics.Append(diags...)

	// Set state
	newState := NewContentTypeData(resource)
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceContentType) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContentTypeData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.ContentTypeDelete(ctx, state.UID.ValueString())
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *resourceContentType) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	resource, err := r.p.stack.ContentTypeUpdate(ctx, state.UID.ValueString(), *input)
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

func (r *resourceContentType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uid"), req, resp)
}

func NewContentTypeData(field *management.ContentType) *ContentTypeData {

	schemaContent, err := field.Schema.MarshalJSON()
	if err != nil {
		panic(err)
	}

	state := &ContentTypeData{
		UID:         types.StringValue(field.UID),
		Title:       types.StringValue(field.Title),
		Description: types.StringValue(field.Description),
		Schema:      types.StringValue(string(schemaContent)),
	}
	return state
}

func NewContentTypeInput(field *ContentTypeData) *management.ContentTypeInput {

	input := &management.ContentTypeInput{
		UID:         field.UID.ValueStringPointer(),
		Title:       field.Title.ValueStringPointer(),
		Description: field.Description.ValueStringPointer(),
		Schema:      json.RawMessage(field.Schema.ValueString()),
	}

	return input
}

func MergeContentType(out *ContentTypeData, in *ContentTypeData) {
	out.Schema = in.Schema
}
