package main

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

// Ensure the implementation satisfies the provider.Provider interface.
var _ provider.Provider = &HLBProvider{}

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/OrographicLift/hlb",
	}
	providerserver.Serve(context.Background(), func() provider.Provider {
		return &HLBProvider{}
	}, opts)
}

// HLBProvider defines the provider implementation.
type HLBProvider struct{}

// HLBProviderModel describes the provider data model.
type HLBProviderModel struct {
	APIKey     types.String `tfsdk:"api_key"`
	AWSRegion  types.String `tfsdk:"aws_region"`
	AWSProfile types.String `tfsdk:"aws_profile"`
	Partition  types.String `tfsdk:"partition"`
}

func (p *HLBProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hlb"
}

func (p *HLBProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with HLB (Hero Load Balancer).",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Description: "API key for HLB authentication. Can also be set with the HLB_API_KEY environment variable.",
				Required:    true,
				Sensitive:   true,
			},
			"aws_region": schema.StringAttribute{
				Description: "AWS region. Can also be set with the AWS_REGION environment variable.",
				Optional:    true,
			},
			"aws_profile": schema.StringAttribute{
				Description: "AWS profile name. Can also be set with the AWS_PROFILE environment variable.",
				Optional:    true,
			},
			"partition": schema.StringAttribute{
				Description: "AWS partition to use. Defaults to 'aws'.",
				Optional:    true,
			},
		},
	}
}

func (p *HLBProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config HLBProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults for optional fields
	if config.Partition.IsNull() {
		config.Partition = types.StringValue("aws")
	}

	// Configure AWS SDK
	var awsOpts []func(*awsconfig.LoadOptions) error
	if !config.AWSProfile.IsNull() {
		awsOpts = append(awsOpts, awsconfig.WithSharedConfigProfile(config.AWSProfile.ValueString()))
	}
	if !config.AWSRegion.IsNull() {
		awsOpts = append(awsOpts, awsconfig.WithRegion(config.AWSRegion.ValueString()))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsOpts...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Configure AWS SDK",
			fmt.Sprintf("Error configuring AWS SDK: %v", err),
		)
		return
	}

	// Create HLB client
	client, err := hlb.NewClient(ctx, config.APIKey.ValueString(), awsCfg, config.Partition.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create HLB Client",
			fmt.Sprintf("Error creating HLB client: %v", err),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *HLBProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewLoadBalancerResource,
		NewListenerAttachmentResource,
	}
}

func (p *HLBProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
