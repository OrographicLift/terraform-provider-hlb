package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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

// preferredMaintenanceWindowValidator validates the format of the preferred maintenance window
type preferredMaintenanceWindowValidator struct{}

// loadBalancerResourceModel maps the resource schema data.
type loadBalancerResourceModel struct {
	AccessLogs                   *accessLogsModel   `tfsdk:"access_logs"`
	ClientKeepAlive              types.Int64        `tfsdk:"client_keep_alive"`
	ConnectionDrainingTimeout    types.Int64        `tfsdk:"connection_draining_timeout"`
	DNSName                      types.String       `tfsdk:"dns_name"`
	Ec2IamRole                   types.String       `tfsdk:"ec2_iam_role"`
	EnableCrossZoneLoadBalancing types.String       `tfsdk:"enable_cross_zone_load_balancing"`
	EnableDeletionProtection     types.Bool         `tfsdk:"enable_deletion_protection"`
	EnableHttp2                  types.Bool         `tfsdk:"enable_http2"`
	ID                           types.String       `tfsdk:"id"`
	IdleTimeout                  types.Int64        `tfsdk:"idle_timeout"`
	Internal                     types.Bool         `tfsdk:"internal"`
	IPAddressType                types.String       `tfsdk:"ip_address_type"`
	LaunchConfig                 *launchConfigModel `tfsdk:"launch_config"`
	Name                         types.String       `tfsdk:"name"`
	NamePrefix                   types.String       `tfsdk:"name_prefix"`
	PreferredMaintenanceWindow   types.String       `tfsdk:"preferred_maintenance_window"`
	PreserveHostHeader           types.Bool         `tfsdk:"preserve_host_header"`
	SecurityGroups               types.Set          `tfsdk:"security_groups"`
	State                        types.String       `tfsdk:"state"`
	Subnets                      types.Set          `tfsdk:"subnets"`
	Tags                         types.Map          `tfsdk:"tags"`
	XffHeaderProcessingMode      types.String       `tfsdk:"xff_header_processing_mode"`
	ZoneID                       types.String       `tfsdk:"zone_id"`
	ZoneName                     types.String       `tfsdk:"zone_name"`
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
		MarkdownDescription: "Manages a Hero Load Balancer (HLB) resource. HLB is a managed load balancer service that " +
			"distributes incoming application traffic across multiple targets, such as EC2 instances, in multiple Availability Zones. " +
			"This resource allows you to create and manage load balancers with features like:\n\n" +
			"* HTTP/2 support\n" +
			"* Access logging\n" +
			"* Cross-zone load balancing\n" +
			"* Custom launch configurations\n" +
			"* Flexible IP address types\n\n" +
			"For detailed documentation and examples, see the [HLB Provider Documentation](../index.md).",
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
				MarkdownDescription: "The name of the load balancer. Must be unique within your account, max 32 characters, " +
					"alphanumeric and hyphens only. This name will be used in DNS records and AWS resource names, so choose " +
					"something meaningful and compliant with DNS naming conventions.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"name_prefix": schema.StringAttribute{
				Optional:    true,
				Description: "Creates a unique name beginning with the specified prefix",
				MarkdownDescription: "Creates a unique name beginning with the specified prefix. Use this when you want " +
					"Terraform to generate unique names for multiple load balancers. The prefix must be 1-6 characters long. " +
					"Conflicts with the `name` attribute - you can only use one or the other.",
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
				MarkdownDescription: "If true, the load balancer will be internal (private). Internal load balancers " +
					"route traffic within your VPC and are not accessible from the internet. Use this for internal " +
					"services or when you want to restrict access to your VPC. This cannot be changed after creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"subnets": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of subnet IDs to attach to the LB",
				MarkdownDescription: "List of subnet IDs to attach to the load balancer. The subnets must be in at " +
					"least two different Availability Zones for high availability. For internal load balancers, these " +
					"must be private subnets. For internet-facing load balancers, these must be public subnets.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"security_groups": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of security group IDs to assign to the LB",
				MarkdownDescription: "List of security group IDs to assign to the load balancer. The security groups " +
					"must allow inbound traffic on the listener ports and health check ports. They should also allow " +
					"outbound traffic to your target instances or IP addresses, and the HLB control plane.",
			},
			"access_logs": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Access logs configuration",
				MarkdownDescription: "Configuration block for access log settings. Access logs capture detailed information about " +
					"requests sent to your load balancer. Each log contains information such as the time the request was received, " +
					"client's IP address, latencies, request paths, and server responses. This can be useful for:\n\n" +
					"* Analyzing traffic patterns\n" +
					"* Troubleshooting issues\n" +
					"* Security analysis\n\n" +
					"The logs are stored in the specified S3 bucket and can be analyzed using tools like Amazon Athena.",
				Attributes: map[string]schema.Attribute{
					"bucket": schema.StringAttribute{
						Required:    true,
						Description: "S3 bucket name",
						MarkdownDescription: "The name of the S3 bucket where access logs will be stored. The load balancer " +
							"must have permissions to write to this bucket. You can use an existing bucket or create a new one. " +
							"Make sure the bucket policy allows the load balancer to write logs.",
					},
					"prefix": schema.StringAttribute{
						Optional:    true,
						Description: "S3 bucket prefix",
						MarkdownDescription: "The prefix (logical hierarchy) in the S3 bucket under which the logs are stored. " +
							"If not specified, logs are stored in the root of the bucket. Example: `my-app/prod/` will store " +
							"logs under that directory structure in the bucket.",
					},
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether to enable access logs",
					},
				},
			},
			"launch_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Launch configuration for the load balancer",
				MarkdownDescription: "Configuration block for customizing how the load balancer instances are launched and scaled. " +
					"This allows you to control the instance type, count, and scaling behavior. If not specified, the backend will " +
					"use default values. It is recommended that you customize these settings to get the best performance / cost ratio," +
					"The best parameters can usually only be determined empirically with a righ-sizing exercise.",
				Attributes: map[string]schema.Attribute{
					"instance_type": schema.StringAttribute{
						Optional:    true,
						Description: "EC2 instance type",
						MarkdownDescription: "The EC2 instance type to use for the load balancer nodes. If not specified, " +
							"the backend will choose an appropriate default (currently c7g.medium). Choose a larger instance " +
							"type if you expect high traffic volumes or need more CPU/memory resources.",
					},
					"min_instance_count": schema.Int64Attribute{
						Optional:    true,
						Description: "Minimum number of instances",
						MarkdownDescription: "The minimum number of instances to maintain (integer, minimum 1). This will be " +
							"rounded up to a multiple of the number of Availability Zones to ensure high availability. If not " +
							"specified, the backend will choose an set this value to the number of AZs.",
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					"max_instance_count": schema.Int64Attribute{
						Optional:    true,
						Description: "Maximum number of instances",
						MarkdownDescription: "The maximum number of instances to maintain (integer, minimum 1). This will be " +
							"rounded down to a multiple of the number of Availability Zones. If not specified, the backend will " +
							"choose 10 times the minimum instance count. This allows for substantial traffic spikes.",
						Validators: []validator.Int64{
							int64validator.AtLeast(1),
						},
					},
					"target_cpu_usage": schema.Int64Attribute{
						Optional:    true,
						Description: "Target CPU usage percentage",
						MarkdownDescription: "The target CPU usage percentage (10-90) for auto-scaling. The load balancer will " +
							"add or remove instances to maintain this target CPU usage. Lower values result in more aggressive " +
							"scaling out, while higher values optimize resource usage but may impact performance during traffic spikes.",
						Validators: []validator.Int64{
							int64validator.Between(10, 90),
						},
					},
				},
			},
			"ec2_iam_role": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(hlb.LBEc2IamRoleStandard),
				Description: "IAM Role to assign to proxy instances. Changing this value exposes you to security risks!",
				MarkdownDescription: "Allows the user to assign a role permitting root access to proxy instances via AWS SSM. " +
					"It is *stronly discouraged* to change this value as this compromises security on many levels. Use only to perfom " +
					"audits. Valid values are:\n" +
					"* `" + hlb.LBEc2IamRoleStandard + "` (default)\n" +
					"* `" + hlb.LBEc2IamRoleDebug + "`\n" +
					"The role will only be applied to newly created proxy instances, you must trigger an Autoscaling Group Instance " +
					"Refresh manually for this change to take effect immediately.",
				Validators: []validator.String{
					stringvalidator.OneOf(hlb.LBEc2IamRoleStandard, hlb.LBEc2IamRoleDebug),
				},
			},
			"enable_deletion_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, deletion of the load balancer will be disabled",
				MarkdownDescription: "If true, deletion of the load balancer will be disabled via the API. This prevents " +
					"accidental deletion of production load balancers. You must set this to false and apply the change " +
					"before you can delete the load balancer.",
			},
			"enable_http2": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Indicates whether HTTP/2 is enabled",
				MarkdownDescription: "Indicates whether HTTP/2 is enabled. HTTP/2 provides improved performance through " +
					"features like multiplexing, header compression, and server push. Enabled by default. Only disable " +
					"this if you have specific compatibility requirements with HTTP/1.x.",
			},
			"idle_timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "The time in seconds that the connection is allowed to be idle",
				MarkdownDescription: "The time in seconds that a connection is allowed to be idle (no data transfer) before " +
					"it is closed by the load balancer. Valid values are 1-4000 seconds. Default is 60 seconds. Increase this " +
					"if your application requires longer-lived idle connections.",
				Validators: []validator.Int64{
					int64validator.Between(1, 4000),
				},
			},
			"ip_address_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ipv4"),
				Description: "The type of IP addresses used by the subnets for your load balancer",
				MarkdownDescription: "The type of IP addresses used by the subnets for your load balancer. Valid values:\n\n" +
					"* `ipv4` (default) - The load balancer has IPv4 addresses only\n" +
					"* `dualstack` - The load balancer has both IPv4 and IPv6 addresses\n" +
					"* `dualstack-without-public-ipv4` - The load balancer has IPv6 addresses and private IPv4 addresses\n\n" +
					"Choose based on your application's networking requirements and client compatibility needs.",
				Validators: []validator.String{
					stringvalidator.OneOf("ipv4", "dualstack", "dualstack-without-public-ipv4"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preserve_host_header": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, preserve the Host header in forwarded requests",
				MarkdownDescription: "If true, the load balancer preserves the original Host header value in requests " +
					"forwarded to targets. This is useful when your backend application needs to know the original hostname " +
					"requested by the client, for example with virtual hosting or when generating absolute URLs.",
			},
			"enable_cross_zone_load_balancing": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("avoid"),
				Description: "Cross-zone load balancing mode",
				MarkdownDescription: "Configures how the load balancer distributes traffic between Availability Zones. Valid values:\n\n" +
					"* `full` - Distributes traffic evenly across all instances in all zones\n" +
					"* `avoid` (default) - Attempts to distribute traffic within a zone first, only sending cross-zone traffic when necessary\n" +
					"* `off` - Keeps traffic within the same zone\n\n" +
					"The `avoid` mode helps optimize for both availability and performance while minimizing cross-AZ data transfer costs.",
				Validators: []validator.String{
					stringvalidator.OneOf("full", "avoid", "off"),
				},
			},
			"client_keep_alive": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3600),
				Description: "The time in seconds to keep client connections alive",
				MarkdownDescription: "The time in seconds to keep client connections alive (60-604800 seconds, default 3600). " +
					"This setting determines how long the load balancer maintains idle keepalive connections with clients. " +
					"Longer values can improve performance by reducing connection establishment overhead, but consume more resources.",
				Validators: []validator.Int64{
					int64validator.Between(60, 604800),
				},
			},
			"xff_header_processing_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("append"),
				Description: "X-Forwarded-For header processing mode",
				MarkdownDescription: "Controls how the load balancer handles the X-Forwarded-For header. Valid values:\n\n" +
					"* `append` (default) - Appends the client IP to any existing X-Forwarded-For header\n" +
					"* `preserve` - Keeps the original X-Forwarded-For header unchanged\n" +
					"* `remove` - Removes any existing X-Forwarded-For header\n\n" +
					"This header is important for identifying the original client IP address when requests pass through the load balancer.",
				Validators: []validator.String{
					stringvalidator.OneOf("append", "preserve", "remove"),
				},
			},
			"tags": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "A map of tags to assign to the resource",
				MarkdownDescription: "A map of tags to assign to the load balancer and related resources. Tags are key-value " +
					"pairs that help you organize and categorize your resources. Common uses include:\n\n" +
					"* Environment identification (e.g., `environment = \"production\"`)\n" +
					"* Cost allocation (e.g., `project = \"mobile-app\"`)\n" +
					"* Security/compliance labeling\n" +
					"* Automation and operations management\n" +
					"A maximum of 5 tags can be assigned (AWS Marketplace limitation).",
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
			"connection_draining_timeout": schema.Int64Attribute{
				Default:     int64default.StaticInt64(10),
				Computed:    true,
				Optional:    true,
				Description: "Load balancer nodes connection draining time in minutes",
				MarkdownDescription: "When a scale-in event occurs, load balancer nodes are taken out of the DNS record immediately " +
					"but we wait for `connection_draining_timeout` minutes before terminating the instances. This is most useful with " +
					"HTTP clients that do not respect DNS TLL.",
				Validators: []validator.Int64{
					int64validator.Between(0, 120),
				},
			},
			"preferred_maintenance_window": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Maintenance window specification in format day-range,hour-range in UTC (e.g., 'mon-wed,02:00-04:00'). Empty string is a valid value.",
				MarkdownDescription: "Specifies a time window during which load balancer node forced replacement should preferably occur. An empty string is a valid value. The format is `ddd-ddd,hh24:mm-hh24:mm` where:\n\n" +
					"* `ddd-ddd` represents a day range (e.g., `mon-fri` for Monday to Friday)\n" +
					"* `hh24:mm-hh24:mm` represents a time range in 24-hour format in UTC (e.g., `01:00-03:00` for 1 AM to 3 AM UTC)\n\n" +
					"**Important:** All times are in UTC. The day of the week refers to the start time of the window. For time ranges that cross midnight, the window will finish on the next day.\n\n" +
					"Examples:\n" +
					"* `mon-wed,02:00-04:00` - Monday through Wednesday, between 2 AM and 4 AM UTC\n" +
					"* `mon-fri,23:00-02:00` - Monday through Friday, starting at 11 PM UTC each day and ending at 2 AM UTC the following day. The last window starts at 11 PM UTC Friday and ends at 2 AM UTC Saturday\n\n" +
					"When updates to the load balancer nodes are available, newly started instances always use the newest version, " +
					"and older versions are progressively phased out as scale-in events occur. Additionally, any remaining instances running " +
					"out-of-date versions will be automatically replaced during the specified maintenance window.\n\n" +
					"Valid values for days are:\n\n" +
					"* `mon` - Monday\n" +
					"* `tue` - Tuesday\n" +
					"* `wed` - Wednesday\n" +
					"* `thu` - Thursday\n" +
					"* `fri` - Friday\n" +
					"* `sat` - Saturday\n" +
					"* `sun` - Sunday\n\n" +
					"**Note:** Urgent updates may be applied outside of the preferred maintenance window and without notice. `connection_draining_timeout` is always respected.\n\n" +
					"If unspecified, non-urgent forced replacements will occur at night in the region.",
				Validators: []validator.String{
					preferredMaintenanceWindowValidatorFunc(),
				},
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
		ClientKeepAlive:              int(plan.ClientKeepAlive.ValueInt64()),
		ConnectionDrainingTimeout:    int(plan.ConnectionDrainingTimeout.ValueInt64()),
		PreferredMaintenanceWindow:   plan.PreferredMaintenanceWindow.ValueString(),
		Ec2IamRole:                   plan.Ec2IamRole.ValueString(),
		EnableCrossZoneLoadBalancing: plan.EnableCrossZoneLoadBalancing.ValueString(),
		EnableDeletionProtection:     plan.EnableDeletionProtection.ValueBool(),
		EnableHttp2:                  plan.EnableHttp2.ValueBool(),
		IdleTimeout:                  int(plan.IdleTimeout.ValueInt64()),
		Internal:                     plan.Internal.ValueBool(),
		IPAddressType:                plan.IPAddressType.ValueString(),
		Name:                         plan.Name.ValueString(),
		NamePrefix:                   plan.NamePrefix.ValueString(),
		PreserveHostHeader:           plan.PreserveHostHeader.ValueBool(),
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
	input.AccessLogs = accessLogsToAPI(plan.AccessLogs)

	// Convert launch config if present
	input.LaunchConfig = launchConfigToAPI(plan.LaunchConfig)

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
	state.ClientKeepAlive = types.Int64Value(int64(lb.ClientKeepAlive))
	state.DNSName = types.StringValue(lb.DNSName)
	state.Ec2IamRole = types.StringValue(lb.Ec2IamRole)
	state.EnableCrossZoneLoadBalancing = types.StringValue(lb.EnableCrossZoneLoadBalancing)
	state.EnableDeletionProtection = types.BoolValue(lb.EnableDeletionProtection)
	state.EnableHttp2 = types.BoolValue(lb.EnableHttp2)
	state.IdleTimeout = types.Int64Value(int64(lb.IdleTimeout))
	state.Internal = types.BoolValue(lb.Internal)
	state.IPAddressType = types.StringValue(lb.IPAddressType)
	state.Name = types.StringValue(lb.Name)
	// Handle PreferredMaintenanceWindow - if it's an empty string from the API, keep the existing value
	if lb.PreferredMaintenanceWindow == "" && !state.PreferredMaintenanceWindow.IsNull() {
		// Keep the existing value
	} else {
		state.PreferredMaintenanceWindow = types.StringValue(lb.PreferredMaintenanceWindow)
	}
	state.PreserveHostHeader = types.BoolValue(lb.PreserveHostHeader)
	state.State = types.StringValue(lb.State)
	state.XffHeaderProcessingMode = types.StringValue(lb.XffHeaderProcessingMode)
	state.ZoneID = types.StringValue(lb.ZoneID)
	state.ZoneName = types.StringValue(lb.ZoneName)

	// Handle AccessLogs - set to null if not present in API response
	if lb.AccessLogs != nil {
		state.AccessLogs = &accessLogsModel{
			Bucket:  types.StringValue(lb.AccessLogs.Bucket),
			Prefix:  types.StringValue(lb.AccessLogs.Prefix),
			Enabled: types.BoolValue(lb.AccessLogs.Enabled),
		}
	} else {
		state.AccessLogs = nil
	}

	// Update LaunchConfig in state if present in the API response
	if lb.LaunchConfig != nil {
		state.LaunchConfig = &launchConfigModel{
			InstanceType:     types.StringValue(lb.LaunchConfig.InstanceType),
			MinInstanceCount: types.Int64Value(int64(lb.LaunchConfig.MinInstanceCount)),
			MaxInstanceCount: types.Int64Value(int64(lb.LaunchConfig.MaxInstanceCount)),
			TargetCPUUsage:   types.Int64Value(int64(lb.LaunchConfig.TargetCPUUsage)),
		}
	} else {
		state.LaunchConfig = nil
	}

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

	if !plan.Ec2IamRole.Equal(state.Ec2IamRole) {
		ec2IamRole := plan.Ec2IamRole.ValueString()
		input.Ec2IamRole = &ec2IamRole
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

	if !plan.ConnectionDrainingTimeout.Equal(state.ConnectionDrainingTimeout) {
		connectionDrainingTimeout := int(plan.ConnectionDrainingTimeout.ValueInt64())
		input.ConnectionDrainingTimeout = &connectionDrainingTimeout
	}

	// Always set PreferredMaintenanceWindow, even if it's not explicitly set in the configuration
	// This ensures that the API always receives the correct value (empty string by default)
	preferredMaintenanceWindow := plan.PreferredMaintenanceWindow.ValueString()
	input.PreferredMaintenanceWindow = &preferredMaintenanceWindow

	// Check for AccessLogs changes
	if plan.AccessLogs != nil {
		input.AccessLogs = accessLogsToAPI(plan.AccessLogs)
	}

	// Check for LaunchConfig changes
	if plan.LaunchConfig != nil {
		input.LaunchConfig = launchConfigToAPI(plan.LaunchConfig)
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

// accessLogsToAPI converts an accessLogsModel to an API AccessLogs object
func accessLogsToAPI(logs *accessLogsModel) *hlb.AccessLogs {
	if logs == nil {
		return nil
	}
	return &hlb.AccessLogs{
		Bucket:  logs.Bucket.ValueString(),
		Prefix:  logs.Prefix.ValueString(),
		Enabled: logs.Enabled.ValueBool(),
	}
}

// launchConfigToAPI converts a launchConfigModel to an API LaunchConfig object
func launchConfigToAPI(config *launchConfigModel) *hlb.LaunchConfig {
	if config == nil {
		return nil
	}
	return &hlb.LaunchConfig{
		InstanceType:     config.InstanceType.ValueString(),
		MinInstanceCount: int(config.MinInstanceCount.ValueInt64()),
		MaxInstanceCount: int(config.MaxInstanceCount.ValueInt64()),
		TargetCPUUsage:   int(config.TargetCPUUsage.ValueInt64()),
	}
}

// Description returns a plain text description of the validator's behavior.
func (v preferredMaintenanceWindowValidator) Description(ctx context.Context) string {
	return "string must be in the format 'ddd-ddd,hh24:mm-hh24:mm' (e.g., 'mon-fri,22:00-02:00') or an empty string"
}

// MarkdownDescription returns a markdown formatted description of the validator's behavior.
func (v preferredMaintenanceWindowValidator) MarkdownDescription(ctx context.Context) string {
	return "string must be in the format `ddd-ddd,hh24:mm-hh24:mm` (e.g., `mon-fri,22:00-02:00`) or an empty string"
}

// ValidateString performs the validation.
func (v preferredMaintenanceWindowValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Empty string is a valid value
	if value == "" {
		return
	}

	// Split into day range and time range
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Maintenance Window Format",
			"The preferred maintenance window must be in the format 'ddd-ddd,hh24:mm-hh24:mm'.",
		)
		return
	}

	dayRange := parts[0]
	timeRange := parts[1]

	// Validate day range
	dayParts := strings.Split(dayRange, "-")
	if len(dayParts) != 2 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Day Range Format",
			"The day range must be in the format 'ddd-ddd' (e.g., 'mon-fri').",
		)
		return
	}

	validDays := map[string]bool{
		"mon": true,
		"tue": true,
		"wed": true,
		"thu": true,
		"fri": true,
		"sat": true,
		"sun": true,
	}

	startDay := strings.ToLower(dayParts[0])
	endDay := strings.ToLower(dayParts[1])

	if !validDays[startDay] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Start Day",
			fmt.Sprintf("'%s' is not a valid day. Valid days are: mon, tue, wed, thu, fri, sat, sun.", startDay),
		)
		return
	}

	if !validDays[endDay] {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid End Day",
			fmt.Sprintf("'%s' is not a valid day. Valid days are: mon, tue, wed, thu, fri, sat, sun.", endDay),
		)
		return
	}

	// Validate time range
	timeRegex := regexp.MustCompile(`^([0-1][0-9]|2[0-3]):([0-5][0-9])-([0-1][0-9]|2[0-3]):([0-5][0-9])$`)
	if !timeRegex.MatchString(timeRange) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Time Range Format",
			"The time range must be in the format 'hh24:mm-hh24:mm' with hours between 00-23 and minutes between 00-59.",
		)
		return
	}
}

// preferredMaintenanceWindowValidatorFunc returns a new preferredMaintenanceWindowValidator.
func preferredMaintenanceWindowValidatorFunc() validator.String {
	return preferredMaintenanceWindowValidator{}
}
