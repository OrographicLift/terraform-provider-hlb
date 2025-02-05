package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &loadBalancerResource{}
	_ resource.ResourceWithConfigure   = &loadBalancerResource{}
	_ resource.ResourceWithImportState = &loadBalancerResource{}
)

// NewLoadBalancerResource is a helper function to simplify the provider implementation.
func NewLoadBalancerResource() resource.Resource {
	return &loadBalancerResource{}
}

// loadBalancerResource is the resource implementation.
type loadBalancerResource struct {
	client *hlb.Client
}

// loadBalancerResourceModel maps the resource schema data.
type loadBalancerResourceModel struct {
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	NamePrefix                   types.String `tfsdk:"name_prefix"`
	Internal                     types.Bool   `tfsdk:"internal"`
	Subnets                      types.Set    `tfsdk:"subnets"`
	SecurityGroups               types.Set    `tfsdk:"security_groups"`
	AccessLogs                   types.List   `tfsdk:"access_logs"`
	LaunchConfig                 types.List   `tfsdk:"launch_config"`
	EnableDeletionProtection     types.Bool   `tfsdk:"enable_deletion_protection"`
	EnableHttp2                  types.Bool   `tfsdk:"enable_http2"`
	IdleTimeout                  types.Int64  `tfsdk:"idle_timeout"`
	IPAddressType                types.String `tfsdk:"ip_address_type"`
	PreserveHostHeader           types.Bool   `tfsdk:"preserve_host_header"`
	EnableCrossZoneLoadBalancing types.String `tfsdk:"enable_cross_zone_load_balancing"`
	ClientKeepAlive              types.Int64  `tfsdk:"client_keep_alive"`
	XffHeaderProcessingMode      types.String `tfsdk:"xff_header_processing_mode"`
	Tags                         types.Map    `tfsdk:"tags"`
	ZoneID                       types.String `tfsdk:"zone_id"`
	ZoneName                     types.String `tfsdk:"zone_name"`
	DNSName                      types.String `tfsdk:"dns_name"`
	State                        types.String `tfsdk:"state"`
}

// accessLogsModel maps the access logs nested object data
type accessLogsModel struct {
	Bucket  types.String `tfsdk:"bucket"`
	Prefix  types.String `tfsdk:"prefix"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

// launchConfigModel maps the launch configuration nested object data
type launchConfigModel struct {
	InstanceType     types.String `tfsdk:"instance_type"`
	MinInstanceCount types.Int64  `tfsdk:"min_instance_count"`
	MaxInstanceCount types.Int64  `tfsdk:"max_instance_count"`
	TargetCPUUsage   types.Int64  `tfsdk:"target_cpu_usage"`
}

// Configure adds the provider configured client to the resource.
func (r *loadBalancerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*hlb.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *hlb.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *loadBalancerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer"
}

// Schema defines the schema for the resource.
func (r *loadBalancerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an HLB Load Balancer",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the load balancer",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"name_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "Creates a unique name beginning with the specified prefix",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 6),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, the LB will be internal",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of subnet IDs to attach to the LB",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"security_groups": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of security group IDs to assign to the LB",
			},
			"access_logs": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Access logs configuration",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"bucket": schema.StringAttribute{
							Required:    true,
							Description: "S3 bucket name",
						},
						"prefix": schema.StringAttribute{
							Optional:    true,
							Description: "S3 bucket prefix",
						},
						"enabled": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Whether to enable access logs",
						},
					},
				},
			},
			"launch_config": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Launch configuration for the load balancer",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"instance_type": schema.StringAttribute{
							Optional:    true,
							Description: "EC2 instance type",
						},
						"min_instance_count": schema.Int64Attribute{
							Optional:    true,
							Description: "Minimum number of instances",
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"max_instance_count": schema.Int64Attribute{
							Optional:    true,
							Description: "Maximum number of instances",
							Validators: []validator.Int64{
								int64validator.AtLeast(1),
							},
						},
						"target_cpu_usage": schema.Int64Attribute{
							Optional:    true,
							Description: "Target CPU usage percentage",
							Validators: []validator.Int64{
								int64validator.Between(10, 90),
							},
						},
					},
				},
			},
			"enable_deletion_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, deletion of the load balancer will be disabled",
			},
			"enable_http2": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Indicates whether HTTP/2 is enabled",
			},
			"idle_timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "The time in seconds that the connection is allowed to be idle",
				Validators: []validator.Int64{
					int64validator.Between(1, 4000),
				},
			},
			"ip_address_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ipv4"),
				Description: "The type of IP addresses used by the subnets for your load balancer",
				Validators: []validator.String{
					stringvalidator.OneOf("ipv4", "dualstack", "dualstack-without-public-ipv4"),
				},
			},
			"preserve_host_header": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, preserve the Host header in forwarded requests",
			},
			"enable_cross_zone_load_balancing": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("avoid"),
				Description: "Cross-zone load balancing mode",
				Validators: []validator.String{
					stringvalidator.OneOf("full", "avoid", "off"),
				},
			},
			"client_keep_alive": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
				Description: "The time in seconds to keep client connections alive",
				Validators: []validator.Int64{
					int64validator.Between(60, 604800),
				},
			},
			"xff_header_processing_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("append"),
				Description: "X-Forwarded-For header processing mode",
				Validators: []validator.String{
					stringvalidator.OneOf("append", "preserve", "remove"),
				},
			},
			"tags": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "A map of tags to assign to the resource",
			},
			"zone_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the zone",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the zone",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dns_name": schema.StringAttribute{
				Computed:    true,
				Description: "The DNS name of the load balancer",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The state of the load balancer",
			},
		},
	}
}

func (r *loadBalancerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan loadBalancerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from plan model to API object
	input := &hlb.LoadBalancerCreate{
		Name:                         plan.Name.ValueString(),
		NamePrefix:                   plan.NamePrefix.ValueString(),
		Internal:                     plan.Internal.ValueBool(),
		EnableDeletionProtection:     plan.EnableDeletionProtection.ValueBool(),
		EnableHttp2:                  plan.EnableHttp2.ValueBool(),
		IdleTimeout:                  int(plan.IdleTimeout.ValueInt64()),
		IPAddressType:                plan.IPAddressType.ValueString(),
		PreserveHostHeader:           plan.PreserveHostHeader.ValueBool(),
		EnableCrossZoneLoadBalancing: plan.EnableCrossZoneLoadBalancing.ValueString(),
		ClientKeepAlive:              int(plan.ClientKeepAlive.ValueInt64()),
		XffHeaderProcessingMode:      plan.XffHeaderProcessingMode.ValueString(),
		ZoneID:                       plan.ZoneID.ValueString(),
		ZoneName:                     plan.ZoneName.ValueString(),
	}

	// Convert subnets from Set to []string
	var subnets []string
	diags = plan.Subnets.ElementsAs(ctx, &subnets, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.Subnets = subnets

	// Convert security groups from Set to []string if present
	if !plan.SecurityGroups.IsNull() {
		var securityGroups []string
		diags = plan.SecurityGroups.ElementsAs(ctx, &securityGroups, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.SecurityGroups = securityGroups
	}

	// Convert access logs if present
	if !plan.AccessLogs.IsNull() {
		var accessLogs []accessLogsModel
		diags = plan.AccessLogs.ElementsAs(ctx, &accessLogs, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(accessLogs) > 0 {
			input.AccessLogs = &hlb.AccessLogs{
				Bucket:  accessLogs[0].Bucket.ValueString(),
				Prefix:  accessLogs[0].Prefix.ValueString(),
				Enabled: accessLogs[0].Enabled.ValueBool(),
			}
		}
	}

	// Convert launch config if present
	if !plan.LaunchConfig.IsNull() {
		var launchConfig []launchConfigModel
		diags = plan.LaunchConfig.ElementsAs(ctx, &launchConfig, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(launchConfig) > 0 {
			input.LaunchConfig = &hlb.LaunchConfig{
				InstanceType:     launchConfig[0].InstanceType.ValueString(),
				MinInstanceCount: int(launchConfig[0].MinInstanceCount.ValueInt64()),
				MaxInstanceCount: int(launchConfig[0].MaxInstanceCount.ValueInt64()),
				TargetCPUUsage:   int(launchConfig[0].TargetCPUUsage.ValueInt64()),
			}
		}
	}

	// Convert tags if present
	if !plan.Tags.IsNull() {
		tags := make(map[string]string)
		diags = plan.Tags.ElementsAs(ctx, &tags, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Tags = tags
	}

	// Create new load balancer
	lb, err := r.client.CreateLoadBalancer(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating load balancer",
			fmt.Sprintf("Could not create load balancer: %v", err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(lb.ID)
	plan.DNSName = types.StringValue(lb.DNSName)
	plan.State = types.StringValue(lb.State)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed load balancer value from HLB
	lb, err := r.client.GetLoadBalancer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HLB Load Balancer",
			fmt.Sprintf("Could not read HLB Load Balancer ID %s: %v", state.ID.ValueString(), err),
		)
		return
	}

	// Overwrite items with refreshed state
	state.DNSName = types.StringValue(lb.DNSName)
	state.State = types.StringValue(lb.State)
	state.Name = types.StringValue(lb.Name)
	state.Internal = types.BoolValue(lb.Internal)
	state.EnableDeletionProtection = types.BoolValue(lb.EnableDeletionProtection)
	state.EnableHttp2 = types.BoolValue(lb.EnableHttp2)
	state.IdleTimeout = types.Int64Value(int64(lb.IdleTimeout))
	state.IPAddressType = types.StringValue(lb.IPAddressType)
	state.PreserveHostHeader = types.BoolValue(lb.PreserveHostHeader)
	state.EnableCrossZoneLoadBalancing = types.StringValue(lb.EnableCrossZoneLoadBalancing)
	state.ClientKeepAlive = types.Int64Value(int64(lb.ClientKeepAlive))
	state.XffHeaderProcessingMode = types.StringValue(lb.XffHeaderProcessingMode)
	state.ZoneID = types.StringValue(lb.ZoneID)
	state.ZoneName = types.StringValue(lb.ZoneName)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan loadBalancerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state loadBalancerResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := &hlb.LoadBalancerUpdate{}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		input.Name = &name
	}

	if !plan.EnableDeletionProtection.Equal(state.EnableDeletionProtection) {
		enableDeletionProtection := plan.EnableDeletionProtection.ValueBool()
		input.EnableDeletionProtection = &enableDeletionProtection
	}

	if !plan.EnableHttp2.Equal(state.EnableHttp2) {
		enableHttp2 := plan.EnableHttp2.ValueBool()
		input.EnableHttp2 = &enableHttp2
	}

	if !plan.IdleTimeout.Equal(state.IdleTimeout) {
		idleTimeout := int(plan.IdleTimeout.ValueInt64())
		input.IdleTimeout = &idleTimeout
	}

	if !plan.PreserveHostHeader.Equal(state.PreserveHostHeader) {
		preserveHostHeader := plan.PreserveHostHeader.ValueBool()
		input.PreserveHostHeader = &preserveHostHeader
	}

	if !plan.EnableCrossZoneLoadBalancing.Equal(state.EnableCrossZoneLoadBalancing) {
		enableCrossZoneLoadBalancing := plan.EnableCrossZoneLoadBalancing.ValueString()
		input.EnableCrossZoneLoadBalancing = &enableCrossZoneLoadBalancing
	}

	if !plan.ClientKeepAlive.Equal(state.ClientKeepAlive) {
		clientKeepAlive := int(plan.ClientKeepAlive.ValueInt64())
		input.ClientKeepAlive = &clientKeepAlive
	}

	if !plan.XffHeaderProcessingMode.Equal(state.XffHeaderProcessingMode) {
		xffHeaderProcessingMode := plan.XffHeaderProcessingMode.ValueString()
		input.XffHeaderProcessingMode = &xffHeaderProcessingMode
	}

	// Update existing load balancer
	_, err := r.client.UpdateLoadBalancer(ctx, state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating HLB Load Balancer",
			fmt.Sprintf("Could not update load balancer %s: %v", state.ID.ValueString(), err),
		)
		return
	}

	// Fetch updated items from GetLoadBalancer as UpdateLoadBalancer doesn't return the full item
	lb, err := r.client.GetLoadBalancer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HLB Load Balancer",
			fmt.Sprintf("Could not read HLB Load Balancer %s: %v", state.ID.ValueString(), err),
		)
		return
	}

	// Update resource state with updated items and timestamp
	plan.ID = types.StringValue(lb.ID)
	plan.DNSName = types.StringValue(lb.DNSName)
	plan.State = types.StringValue(lb.State)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *loadBalancerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state loadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing load balancer
	err := r.client.DeleteLoadBalancer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting HLB Load Balancer",
			fmt.Sprintf("Could not delete load balancer %s: %v", state.ID.ValueString(), err),
		)
		return
	}
}

func (r *loadBalancerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
