package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/contentstack-go-sdk/management"
)

// retryableRoundTripper implements http.RoundTripper using retryablehttp.Client
type retryableRoundTripper struct {
	retryClient *retryablehttp.Client
}

// RoundTrip implements the http.RoundTripper interface
func (r *retryableRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Convert http.Request to retryablehttp.Request
	retryReq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}
	// Use the retryable client's Do method
	return r.retryClient.Do(retryReq)
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{version: version}
	}
}

type provider struct {
	stack   *management.StackInstance
	client  *management.Client
	version string
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
			"branch": {
				Type:        types.StringType,
				Optional:    true,
				Description: "The branch to manage resources in. If not specified, the main branch will be used.",
			},
			"max_retries": {
				Type:        types.Int64Type,
				Optional:    true,
				Description: "The maximum number of retry attempts for 429 (rate limit) responses. Defaults to 3. Uses exponential backoff with jitter.",
			},
			"retry_wait_min": {
				Type:        types.Int64Type,
				Optional:    true,
				Description: "The minimum wait time in seconds between retries. Defaults to 1 second.",
			},
			"retry_wait_max": {
				Type:        types.Int64Type,
				Optional:    true,
				Description: "The maximum wait time in seconds between retries. Defaults to 30 seconds.",
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
	Branch          types.String `tfsdk:"branch"`
	MaxRetries      types.Int64  `tfsdk:"max_retries"`
	RetryWaitMin    types.Int64  `tfsdk:"retry_wait_min"`
	RetryWaitMax    types.Int64  `tfsdk:"retry_wait_max"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	// Retrieve provider data from configuration
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = nil

	maxRetries := 3
	if !config.MaxRetries.IsNull() && !config.MaxRetries.IsUnknown() {
		maxRetries = int(config.MaxRetries.Value)
	}
	retryClient.RetryMax = maxRetries

	retryWaitMin := 1 * time.Second
	if !config.RetryWaitMin.IsNull() && !config.RetryWaitMin.IsUnknown() {
		retryWaitMin = time.Duration(config.RetryWaitMin.Value) * time.Second
	}
	retryClient.RetryWaitMin = retryWaitMin

	retryWaitMax := 30 * time.Second
	if !config.RetryWaitMax.IsNull() && !config.RetryWaitMax.IsUnknown() {
		retryWaitMax = time.Duration(config.RetryWaitMax.Value) * time.Second
	}
	retryClient.RetryWaitMax = retryWaitMax

	// Custom retry policy to only retry on 429 responses
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		// Delegate network/transport errors to the default retry policy
		if err != nil {
			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}
		// Only retry on 429 (rate limit) responses
		if resp.StatusCode == 429 {
			return true, nil
		}
		return false, nil
	}

	// Set the underlying transport to use DebugTransport for logging
	// This maintains the original behavior while adding retry capability
	retryClient.HTTPClient = &http.Client{
		Transport: management.DebugTransport,
	}

	// Create a custom RoundTripper that uses the retryable client
	roundTripper := &retryableRoundTripper{retryClient: retryClient}

	httpClient := &http.Client{
		Transport: roundTripper,
	}

	cfg := management.ClientConfig{
		BaseURL:    config.BaseURL.Value,
		AuthToken:  config.AuthToken.Value,
		HTTPClient: httpClient,
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
		Branch:          config.Branch.Value,
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
		"contentstack_environment":  resourceEnvironmentType{},
		"contentstack_global_field": resourceGlobalFieldType{},
		"contentstack_locale":       resourceLocaleType{},
		"contentstack_webhook":      resourceWebhookType{},
	}, nil
}

// GetDataSources - Defines provider data sources
func (p *provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}
