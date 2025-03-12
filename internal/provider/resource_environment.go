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

type resourceEnvironmentType struct{}

type EnvironmentData struct {
	UID  types.String         `tfsdk:"uid"`
	Name types.String         `tfsfk:"name"`
	URLs []EnvironmentUrlData `tfsdk:"url"`
}

type EnvironmentUrlData struct {
	Locale types.String `tfsdk:"locale"`
	URL    types.String `tfsdk:"url"`
}

// Environment Resource schema
func (r resourceEnvironmentType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Description: `
		Contentstack environment are designated destinations to which you can publish
		your content. Environments are global, meaning they are available across all
		branches of your stack. An environment can also have a list of URLs to be used
		as a prefix for published content.
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
		},
		Blocks: map[string]tfsdk.Block{
			"url": {
				NestingMode: tfsdk.BlockNestingModeList,
				Blocks:      map[string]tfsdk.Block{},
				MinItems:    0,
				Attributes: map[string]tfsdk.Attribute{
					"locale": {
						Type:     types.StringType,
						Required: true,
					},
					"url": {
						Type:     types.StringType,
						Required: true,
					},
				},
			},
		},
	}, nil
}

// New resource instance
func (r resourceEnvironmentType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceEnvironment{
		p: *(p.(*provider)),
	}, nil
}

type resourceEnvironment struct {
	p provider
}

func (r resourceEnvironment) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
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

func (r resourceEnvironment) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state EnvironmentData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environment, err := r.p.stack.EnvironmentFetch(ctx, state.UID.Value)
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving environment",
				fmt.Sprintf("The environment with UID %s was not found.", state.UID.Value))
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

func (r resourceEnvironment) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state EnvironmentData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete environment by calling API
	err := r.p.stack.EnvironmentDelete(ctx, state.UID.Value)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceEnvironment) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
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
	environment, err := r.p.stack.EnvironmentUpdate(ctx, state.UID.Value, *input)
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

func (r resourceEnvironment) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func NewEnvironmentData(environment *management.Environment) *EnvironmentData {
	urls := []EnvironmentUrlData{}
	for i := range environment.URLs {
		s := environment.URLs[i]

		url := EnvironmentUrlData{
			Locale: types.String{Value: s.Locale},
			URL:    types.String{Value: s.URL},
		}

		urls = append(urls, url)
	}

	state := &EnvironmentData{
		UID:  types.String{Value: environment.UID},
		Name: types.String{Value: environment.Name},
		URLs: urls,
	}
	return state
}

func NewEnvironmentInput(environment *EnvironmentData) *management.EnvironmentInput {
	urls := []management.EnvironmentUrl{}
	for i := range environment.URLs {
		s := environment.URLs[i]
		url := management.EnvironmentUrl{
			Locale: s.Locale.Value,
			URL:    s.URL.Value,
		}

		urls = append(urls, url)
	}

	input := &management.EnvironmentInput{
		Name: environment.Name.Value,
		URLs: urls,
	}

	return input
}
