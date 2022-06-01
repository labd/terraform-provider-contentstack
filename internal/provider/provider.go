package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/contentstack-go-sdk/management"
)

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"base_url": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"api_key": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"management_token": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
				"auth_token": {
					Type:      schema.TypeString,
					Optional:  true,
					Sensitive: true,
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"contentstack_webhook": resourceWebhook(),
			},
			DataSourcesMap: map[string]*schema.Resource{},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {

	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		baseURL := d.Get("base_url").(string)
		authToken := d.Get("auth_token").(string)
		apiKey := d.Get("api_key").(string)
		managementToken := d.Get("management_token").(string)

		// Warning or errors can be collected in a slice type
		var diags diag.Diagnostics

		cfg := management.ClientConfig{
			BaseURL:   baseURL,
			AuthToken: authToken,
		}

		c, err := management.NewClient(cfg)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		stackAuth := management.StackAuth{
			ApiKey:          apiKey,
			ManagementToken: managementToken,
		}

		instance, err := c.Stack(&stackAuth)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return instance, diags
	}
}
