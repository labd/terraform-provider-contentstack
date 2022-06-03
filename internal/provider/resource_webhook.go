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

type resourceWebhookType struct{}

type WebhookData struct {
	UID            types.String             `tfsdk:"uid"`
	Name           types.String             `tfsdk:"name"`
	Branches       []types.String           `tfsdk:"branches"`
	Channels       []types.String           `tfsdk:"channels"`
	RetryPolicy    types.String             `tfsdk:"retry_policy"`
	ConcisePayload types.Bool               `tfsdk:"concise_payload"`
	Disabled       types.Bool               `tfsdk:"disabled"`
	Destinations   []WebhookDestinationData `tfsdk:"destination"`
}

type WebhookDestinationData struct {
	TargetURL         types.String              `tfsdk:"target_url"`
	HttpBasicAuth     types.String              `tfsdk:"http_basic_auth"`
	HttpBasicPassword types.String              `tfsdk:"http_basic_password"`
	CustomHeaders     []WebhookCustomHeaderData `tfsdk:"custom_headers"`
}

type WebhookCustomHeaderData struct {
	Name  types.String `tfsdk:"header_name"`
	Value types.String `tfsdk:"value"`
}

// Webhook Resource schema
func (r resourceWebhookType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"uid": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"branches": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"channels": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"retry_policy": {
				Type:     types.StringType,
				Required: true,
			},
			"disabled": {
				Type:     types.BoolType,
				Optional: true,
			},
			"concise_payload": {
				Type:     types.BoolType,
				Optional: true,
			},
		},
		Blocks: map[string]tfsdk.Block{
			"destination": {
				NestingMode: tfsdk.BlockNestingModeList,
				Blocks:      map[string]tfsdk.Block{},
				MinItems:    1,
				Attributes: map[string]tfsdk.Attribute{
					"target_url": {
						Type:     types.StringType,
						Required: true,
					},
					"http_basic_auth": {
						Type:     types.StringType,
						Required: true,
					},
					"http_basic_password": {
						Type:      types.StringType,
						Required:  true,
						Sensitive: true,
					},
					"custom_headers": {
						Optional: true,
						Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
							"header_name": {
								Type:     types.StringType,
								Required: true,
							},
							"value": {
								Type:     types.StringType,
								Required: true,
							},
						}),
					},
				},
			},
		},
	}, nil
}

// New resource instance
func (r resourceWebhookType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceWebhook{
		p: *(p.(*provider)),
	}, nil
}

type resourceWebhook struct {
	p provider
}

func (r resourceWebhook) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var plan WebhookData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewWebhookInput(&plan)
	webhook, err := r.p.stack.WebHookCreate(ctx, *input)
	if err != nil {
		diags := processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(webhook, input)
	resp.Diagnostics.Append(diags...)

	// Write to state
	state := NewWebhookData(webhook)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r resourceWebhook) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state WebhookData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.p.stack.WebHookFetch(ctx, state.UID.Value)
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving webhook",
				fmt.Sprintf("The webhook with UID %s was not found.", state.UID.Value))
			resp.Diagnostics.Append(d)
		} else {
			diags := processRemoteError(err)
			resp.Diagnostics.Append(diags...)
		}
		return
	}

	curr := NewWebhookInput(&state)
	diags = processResponse(webhook, curr)
	resp.Diagnostics.Append(diags...)

	// Set state
	newState := NewWebhookData(webhook)
	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r resourceWebhook) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state WebhookData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.WebHookDelete(ctx, state.UID.Value)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceWebhook) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan WebhookData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state WebhookData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewWebhookInput(&plan)
	webhook, err := r.p.stack.WebHookUpdate(ctx, state.UID.Value, *input)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = processResponse(webhook, input)
	resp.Diagnostics.Append(diags...)

	// Set state
	result := NewWebhookData(webhook)
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r resourceWebhook) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func NewWebhookData(webhook *management.WebHook) *WebhookData {
	branches := []types.String{}
	for i := range webhook.Branches {
		branches = append(branches, types.String{Value: webhook.Branches[i]})
	}

	channels := []types.String{}
	for i := range webhook.Channels {
		channels = append(channels, types.String{Value: webhook.Channels[i]})
	}

	destinations := []WebhookDestinationData{}
	for i := range webhook.Destinations {
		s := webhook.Destinations[i]

		dest := WebhookDestinationData{
			TargetURL:         types.String{Value: s.TargetURL},
			HttpBasicAuth:     types.String{Value: s.HttpBasicAuth},
			HttpBasicPassword: types.String{Value: s.HttpBasicPassword},
		}

		for j := range s.CustomHeaders {
			header := WebhookCustomHeaderData{
				Name:  types.String{Value: s.CustomHeaders[j].Name},
				Value: types.String{Value: s.CustomHeaders[j].Value},
			}
			dest.CustomHeaders = append(dest.CustomHeaders, header)
		}

		destinations = append(destinations, dest)
	}

	state := &WebhookData{
		UID:            types.String{Value: webhook.UID},
		Name:           types.String{Value: webhook.Name},
		RetryPolicy:    types.String{Value: webhook.RetryPolicy},
		ConcisePayload: types.Bool{Value: webhook.ConcisePayload},
		Disabled:       types.Bool{Value: webhook.Disabled},
		Channels:       channels,
		Branches:       branches,
		Destinations:   destinations,
	}
	return state
}

func NewWebhookInput(webhook *WebhookData) *management.WebHookInput {
	destinations := []management.WebhookDestination{}
	for i := range webhook.Destinations {
		s := webhook.Destinations[i]
		dest := management.WebhookDestination{
			TargetURL:         s.TargetURL.Value,
			HttpBasicAuth:     s.HttpBasicAuth.Value,
			HttpBasicPassword: s.HttpBasicPassword.Value,
		}

		for j := range s.CustomHeaders {
			header := management.WebhookHeader{
				Name:  s.CustomHeaders[j].Name.Value,
				Value: s.CustomHeaders[j].Value.Value,
			}
			dest.CustomHeaders = append(dest.CustomHeaders, header)
		}
		destinations = append(destinations, dest)
	}

	input := &management.WebHookInput{
		Name:           webhook.Name.Value,
		RetryPolicy:    webhook.RetryPolicy.Value,
		Destinations:   destinations,
		ConcisePayload: webhook.ConcisePayload.Value,
	}
	for i := range webhook.Channels {
		input.Channels = append(input.Channels, webhook.Channels[i].Value)
	}
	for i := range webhook.Branches {
		input.Branches = append(input.Branches, webhook.Branches[i].Value)
	}

	return input
}
