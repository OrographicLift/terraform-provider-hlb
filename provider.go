package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: Provider,
	})
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("HLB_API_KEY", nil),
			},
			"aws_region": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("AWS_REGION", nil),
			},
			"aws_profile": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc("AWS_PROFILE", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"hlb_load_balancer": resourceHLBLoadBalancer(),
			"hlb_listener_attachment":  resourceHLBListenerAttachment(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	awsRegion := d.Get("aws_region").(string)
	awsProfile := d.Get("aws_profile").(string)

	opts := []func(*config.LoadOptions) error{}
	if awsProfile != "" {
		opts = append(opts, config.WithSharedConfigProfile(awsProfile))
	}
	if awsRegion != "" {
		opts = append(opts, config.WithRegion(awsRegion))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("error configuring AWS SDK: %v", err))
	}

	// Create the HLB client
	client, err := hlb.NewClient(apiKey, awsCfg)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("error creating HLB client: %v", err))
	}

	return client, nil
}