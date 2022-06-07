package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/contentstack-go-sdk/management"
)

var stderr = os.Stderr

func New() tfsdk.Provider {
	return &provider{}
}

type provider struct {
	configured bool
	stack      *management.StackInstance
	client     *management.Client
}

// GetSchema
func (p *provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"base_url": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The BaseURL, e.g. https://eu-api.contentstack.com/. See https://www.contentstack.com/docs/developers/apis/content-management-api/#base-url",
			},
			"api_key": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The API key is a unique key assigned to each stack.",
			},
			"management_token": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "Management Tokens are stack-level tokens, with no users attached to them.",
			},
			"auth_token": {
				Type:        types.StringType,
				Optional:    true,
				Sensitive:   true,
				Description: "The Authtoken is a read-write token used to make authorized CMA requests, and it is a user-specific token.",
			},
		},
	}, nil
}

// Provider schema struct
type providerData struct {
	BaseURL         types.String `tfsdk:"base_url"`
	AuthToken       types.String `tfsdk:"auth_token"`
	ApiKey          types.String `tfsdk:"api_key"`
	ManagementToken types.String `tfsdk:"management_token"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {

	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := management.ClientConfig{
		BaseURL:   config.BaseURL.Value,
		AuthToken: config.AuthToken.Value,
		HTTPClient: &http.Client{
			Transport: management.DebugTransport,
		},
	}

	c, err := management.NewClient(cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create client",
			"Unable to create contentstack client:\n\n"+err.Error(),
		)
		return
	}

	stackAuth := management.StackAuth{
		ApiKey:          config.ApiKey.Value,
		ManagementToken: config.ManagementToken.Value,
	}

	instance, err := c.Stack(&stackAuth)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create stack client",
			"Unable to create contentstack stack client:\n\n"+err.Error(),
		)
		return
	}

	p.client = c
	p.stack = instance
}

// GetResources - Defines provider resources
func (p *provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"contentstack_content_type": resourceContentTypeType{},
		"contentstack_global_field": resourceGlobalFieldType{},
		"contentstack_locale":       resourceLocaleType{},
		"contentstack_webhook":      resourceWebhookType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
