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

type EnvironmentData struct {
	UID  types.String         `tfsdk:"uid"`
	Name types.String         `tfsdk:"name"`
	URLs []EnvironmentUrlData `tfsdk:"url"`
}

type EnvironmentUrlData struct {
	Locale types.String `tfsdk:"locale"`
	URL    types.String `tfsdk:"url"`
}

type resourceEnvironment struct {
	p *contentstackProvider
}

// Metadata
func (r *resourceEnvironment) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "contentstack_environment"
}

// Environment Resource schema
func (r *resourceEnvironment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
		Contentstack environment are designated destinations to which you can publish
		your content. Environments are global, meaning they are available across all
		branches of your stack. An environment can also have a list of URLs to be used
		as a prefix for published content.
		`,
		Attributes: map[string]schema.Attribute{
			"uid": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"url": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"locale": schema.StringAttribute{
							Required: true,
						},
						"url": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceEnvironment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewEnvironmentInput(&plan)
	environment, err := r.p.stack.EnvironmentCreate(ctx, *input)
	if err != nil {
		diags := processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(environment, input)
	resp.Diagnostics.Append(diags...)

	// Write to state
	state := NewEnvironmentData(environment)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceEnvironment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.p.stack.EnvironmentFetch(ctx, state.Name.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving environment",
				fmt.Sprintf("The environment with Name %s was not found.", state.Name.ValueString()))
			resp.Diagnostics.Append(d)
		} else {
			diags := processRemoteError(err)
			resp.Diagnostics.Append(diags...)
		}
		return
	}

	curr := NewEnvironmentInput(&state)
	diags = processResponse(environment, curr)
	resp.Diagnostics.Append(diags...)

	// Set state
	newState := NewEnvironmentData(environment)
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceEnvironment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete environment by calling API
	err := r.p.stack.EnvironmentDelete(ctx, state.Name.ValueString())
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *resourceEnvironment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan values
	var plan EnvironmentData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state EnvironmentData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewEnvironmentInput(&plan)
	environment, err := r.p.stack.EnvironmentUpdate(ctx, state.Name.ValueString(), *input)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(environment, input)
	resp.Diagnostics.Append(diags...)

	// Set state
	result := NewEnvironmentData(environment)
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceEnvironment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uid"), req, resp)
}

func NewEnvironmentData(environment *management.Environment) *EnvironmentData {
	urls := []EnvironmentUrlData{}
	for i := range environment.URLs {
		s := environment.URLs[i]

		url := EnvironmentUrlData{
			Locale: types.StringValue(s.Locale),
			URL:    types.StringValue(s.URL),
		}

		urls = append(urls, url)
	}

	state := &EnvironmentData{
		UID:  types.StringValue(environment.UID),
		Name: types.StringValue(environment.Name),
		URLs: urls,
	}
	return state
}

func NewEnvironmentInput(environment *EnvironmentData) *management.EnvironmentInput {
	urls := []management.EnvironmentUrl{}
	for i := range environment.URLs {
		s := environment.URLs[i]
		url := management.EnvironmentUrl{
			Locale: s.Locale.ValueString(),
			URL:    s.URL.ValueString(),
		}

		urls = append(urls, url)
	}

	input := &management.EnvironmentInput{
		Name: environment.Name.ValueString(),
		URLs: urls,
	}

	return input
}
