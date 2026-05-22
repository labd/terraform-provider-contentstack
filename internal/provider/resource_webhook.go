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

type WebhookData struct {
	UID            types.String            `tfsdk:"uid"`
	Name           types.String            `tfsdk:"name"`
	Branches       []types.String          `tfsdk:"branches"`
	Channels       []types.String          `tfsdk:"channels"`
	RetryPolicy    types.String            `tfsdk:"retry_policy"`
	ConcisePayload types.Bool              `tfsdk:"concise_payload"`
	Disabled       types.Bool              `tfsdk:"disabled"`
	Destinations   WebhookDestinationSlice `tfsdk:"destination"`
}

type WebhookDestinationSlice []WebhookDestinationData

func (s *WebhookDestinationSlice) FindByTargetURLAndHttpBasicAuth(t, a string) *WebhookDestinationData {
	for i := range *s {
		if (*s)[i].TargetURL.ValueString() == t && (*s)[i].HttpBasicAuth.ValueString() == a {
			return &(*s)[i]
		}
	}
	return nil
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

type resourceWebhook struct {
	p *contentstackProvider
}

// Metadata
func (r *resourceWebhook) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "contentstack_webhook"
}

func (r *resourceWebhook) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
		A webhook is a user-defined HTTP callback. It is a mechanism that sends
		real-time information to any third-party app or service.
		`,
		Attributes: map[string]schema.Attribute{
			"uid": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"branches": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"channels": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"retry_policy": schema.StringAttribute{
				Required:    true,
				Description: "should be set to `manual`",
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Description: "allows you to enable or disable the webhook.",
			},
			"concise_payload": schema.BoolAttribute{
				Optional:    true,
				Description: "allows you to send a concise JSON payload to the target URL when a specific event occurs. To send a comprehensive JSON payload, you can set its value to false.",
			},
		},
		Blocks: map[string]schema.Block{
			"destination": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"target_url": schema.StringAttribute{
							Required: true,
						},
						"http_basic_auth": schema.StringAttribute{
							Required: true,
						},
						"http_basic_password": schema.StringAttribute{
							Required:  true,
							Sensitive: true,
						},
						"custom_headers": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"header_name": schema.StringAttribute{
										Required: true,
									},
									"value": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceWebhook) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := NewWebhookInput(&plan)
	webhook, err := r.p.stack.WebHookCreate(ctx, *input)
	if err != nil {
		resp.Diagnostics.Append(processRemoteError(err)...)
		return
	}

	diags = processResponse(webhook, input)
	resp.Diagnostics.Append(diags...)

	//Set actual password as state
	webhook.Destinations, err = copyHttpBasicPasswords(webhook.Destinations, plan.Destinations)
	if err != nil {
		resp.Diagnostics.AddError("Error copying http basic passwords", err.Error())
		return
	}

	// Write to state
	state := NewWebhookData(webhook)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceWebhook) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhook, err := r.p.stack.WebHookFetch(ctx, state.UID.ValueString())
	if err != nil {
		if IsNotFoundError(err) {
			d := diag.NewErrorDiagnostic(
				"Error retrieving webhook",
				fmt.Sprintf("The webhook with UID %s was not found.", state.UID.ValueString()))
			resp.Diagnostics.Append(d)
		} else {
			diags := processRemoteError(err)
			resp.Diagnostics.Append(diags...)
		}
		return
	}

	//Set actual password as state
	webhook.Destinations, err = copyHttpBasicPasswords(webhook.Destinations, state.Destinations)
	if err != nil {
		resp.Diagnostics.AddError("Error copying http basic passwords", err.Error())
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

func (r *resourceWebhook) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete order by calling API
	err := r.p.stack.WebHookDelete(ctx, state.UID.ValueString())
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r *resourceWebhook) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
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
	webhook, err := r.p.stack.WebHookUpdate(ctx, state.UID.ValueString(), *input)
	if err != nil {
		diags = processRemoteError(err)
		resp.Diagnostics.Append(diags...)
		return
	}

	webhook.Destinations, err = copyHttpBasicPasswords(webhook.Destinations, plan.Destinations)
	if err != nil {
		resp.Diagnostics.AddError("Error copying http basic passwords", err.Error())
		return
	}

	diags = processResponse(webhook, input)
	resp.Diagnostics.Append(diags...)

	// Set state
	result := NewWebhookData(webhook)
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceWebhook) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uid"), req, resp)
}

func NewWebhookData(webhook *management.WebHook) *WebhookData {
	var branches []types.String
	for i := range webhook.Branches {
		branches = append(branches, types.StringValue(webhook.Branches[i]))
	}

	var channels []types.String
	for i := range webhook.Channels {
		channels = append(channels, types.StringValue(webhook.Channels[i]))
	}

	var destinations []WebhookDestinationData
	for i := range webhook.Destinations {
		s := webhook.Destinations[i]

		dest := WebhookDestinationData{
			TargetURL:         types.StringValue(s.TargetURL),
			HttpBasicAuth:     types.StringValue(s.HttpBasicAuth),
			HttpBasicPassword: types.StringValue(s.HttpBasicPassword),
		}

		for j := range s.CustomHeaders {
			header := WebhookCustomHeaderData{
				Name:  types.StringValue(s.CustomHeaders[j].Name),
				Value: types.StringValue(s.CustomHeaders[j].Value),
			}
			dest.CustomHeaders = append(dest.CustomHeaders, header)
		}

		destinations = append(destinations, dest)
	}

	state := &WebhookData{
		UID:            types.StringValue(webhook.UID),
		Name:           types.StringValue(webhook.Name),
		RetryPolicy:    types.StringValue(webhook.RetryPolicy),
		ConcisePayload: types.BoolValue(webhook.ConcisePayload),
		Disabled:       types.BoolValue(webhook.Disabled),
		Channels:       channels,
		Branches:       branches,
		Destinations:   destinations,
	}
	return state
}

func NewWebhookInput(webhook *WebhookData) *management.WebHookInput {
	var destinations []management.WebhookDestination
	for i := range webhook.Destinations {
		s := webhook.Destinations[i]
		dest := management.WebhookDestination{
			TargetURL:         s.TargetURL.ValueString(),
			HttpBasicAuth:     s.HttpBasicAuth.ValueString(),
			HttpBasicPassword: s.HttpBasicPassword.ValueString(),
		}

		for j := range s.CustomHeaders {
			header := management.WebhookHeader{
				Name:  s.CustomHeaders[j].Name.ValueString(),
				Value: s.CustomHeaders[j].Value.ValueString(),
			}
			dest.CustomHeaders = append(dest.CustomHeaders, header)
		}
		destinations = append(destinations, dest)
	}

	input := &management.WebHookInput{
		Name:           webhook.Name.ValueString(),
		RetryPolicy:    webhook.RetryPolicy.ValueString(),
		Destinations:   destinations,
		ConcisePayload: webhook.ConcisePayload.ValueBool(),
	}
	for i := range webhook.Channels {
		input.Channels = append(input.Channels, webhook.Channels[i].ValueString())
	}
	for i := range webhook.Branches {
		input.Branches = append(input.Branches, webhook.Branches[i].ValueString())
	}

	return input
}
