package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource              = &listenerAttachmentResource{}
	_ resource.ResourceWithConfigure = &listenerAttachmentResource{}
)

// NewListenerAttachmentResource is a helper function to simplify the provider implementation.
func NewListenerAttachmentResource() resource.Resource {
	return &listenerAttachmentResource{}
}

// listenerAttachmentResource is the resource implementation.
type listenerAttachmentResource struct {
	client *hlb.Client
}

// listenerAttachmentResourceModel maps the resource schema data.
type listenerAttachmentResourceModel struct {
	ID                       types.String  `tfsdk:"id"`
	ALPNPolicy               types.String  `tfsdk:"alpn_policy"`
	CertificateSecretsName   types.String  `tfsdk:"certificate_secrets_name"`
	EnableDeletionProtection types.Bool    `tfsdk:"enable_deletion_protection"`
	LoadBalancerID           types.String  `tfsdk:"load_balancer_id"`
	OverprovisioningFactor   types.Float64 `tfsdk:"overprovisioning_factor"`
	Port                     types.Int64   `tfsdk:"port"`
	Protocol                 types.String  `tfsdk:"protocol"`
	TargetGroupARN           types.String  `tfsdk:"target_group_arn"`
}

// Configure adds the provider configured client to the resource.
func (r *listenerAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *listenerAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_listener_attachment"
}

// Schema defines the schema for the resource.
func (r *listenerAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an HLB Listener Attachment",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"alpn_policy": schema.StringAttribute{
				Optional:    true,
				Description: "Application-Layer Protocol Negotiation (ALPN) policy",
				MarkdownDescription: "The Application-Layer Protocol Negotiation (ALPN) policy to use. Valid values:\n\n" +
					"* `HTTP1Only` - Only allow HTTP/1.* connections\n" +
					"* `HTTP2Only` - Only allow HTTP/2 connections\n" +
					"* `HTTP2Optional` - Prefer HTTP/1.*, but allow HTTP/2 if client supports it\n" +
					"* `HTTP2Preferred` - Prefer HTTP/2, but allow HTTP/1.* if client doesn't support HTTP/2\n" +
					"* `None` - Do not use ALPN",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"HTTP1Only",
						"HTTP2Only",
						"HTTP2Optional",
						"HTTP2Preferred",
						"None",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_secrets_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the certificate secrets",
				MarkdownDescription: "The name of the secret in AWS Secrets Manager containing the SSL server certificate. " +
					"This field is required if the protocol is HTTPS. The secret should contain the certificate and private key in PEM format.",
			},
			"enable_deletion_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to enable deletion protection",
				MarkdownDescription: "If true, deletion of the listener will be disabled via the API. This prevents accidental deletion " +
					"of production listeners. You must set this to false and apply the change before you can delete the listener.",
			},
			"load_balancer_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the load balancer",
				MarkdownDescription: "The ID of the HLB load balancer to attach this listener to. This cannot be changed once the listener " +
					"is created - you must create a new listener to move it to a different load balancer.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"overprovisioning_factor": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(1.1),
				Description: "Overprovisioning factor for the listener",
				MarkdownDescription: "When the load balancer `enable_cross_zone_load_balancing` is set to `avoid` (default), this factor " +
					"determines how much traffic a single instance can receive relative to the average. For example, with a value of `1.1`, " +
					"an instance can receive up to 110% of the average traffic. This helps optimize resource usage while maintaining zone isolation. " +
					"Must be greater than or equal to 1.0. Applies to the target group - behavior is undefined if different values are assigned " +
					"to different listeners with the same target group.",
				Validators: []validator.Float64{
					float64validator.AtLeast(1.0),
				},
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "Port on which the listener accepts connection",
				MarkdownDescription: "The port on which the load balancer is listening. Must be between 1 and 65535. Common ports:\n\n" +
					"* 80 - HTTP\n" +
					"* 443 - HTTPS\n\n" +
					"Note: The security groups associated with the load balancer must allow inbound traffic on this port.",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "Protocol for connections from clients to the listener",
				MarkdownDescription: "The protocol for connections from clients to the load balancer. Valid values:\n\n" +
					"* `HTTP` - For unencrypted HTTP traffic\n" +
					"* `HTTPS` - For encrypted HTTPS traffic. When using HTTPS, you must also specify a `certificate_secrets_name`\n\n" +
					"* `UDP` - For raw unencrypted UDP traffic\n\n" +
					"The protocol cannot be changed after the listener is created - you must create a new listener to change protocols.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"HTTP",
						"HTTPS",
						"UDP",
					),
				},
			},
			"target_group_arn": schema.StringAttribute{
				Required:    true,
				Description: "ARN of the target group",
				MarkdownDescription: "The Amazon Resource Name (ARN) of the target group to route traffic to. The target group defines " +
					"where traffic will be sent (the backend instances, IP addresses, or other resources) and how health checks are performed. " +
					"The target group must be in the same region as the load balancer.",
			},
		},
	}
}

func (r *listenerAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan listenerAttachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create new listener
	input := &hlb.ListenerCreate{
		EnableDeletionProtection: plan.EnableDeletionProtection.ValueBool(),
		OverprovisioningFactor:   plan.OverprovisioningFactor.ValueFloat64(),
		Port:                     int(plan.Port.ValueInt64()),
		Protocol:                 plan.Protocol.ValueString(),
		TargetGroupARN:           plan.TargetGroupARN.ValueString(),
	}

	// Only set optional string fields if they have a value
	if !plan.ALPNPolicy.IsNull() {
		input.ALPNPolicy = plan.ALPNPolicy.ValueString()
	}
	if !plan.CertificateSecretsName.IsNull() {
		input.CertificateSecretsName = plan.CertificateSecretsName.ValueString()
	}

	listener, err := r.client.CreateListener(ctx, plan.LoadBalancerID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating listener",
			fmt.Sprintf("Could not create listener: %v", err),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(listener.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *listenerAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state listenerAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed listener value from HLB
	listener, err := r.client.GetListener(ctx, state.LoadBalancerID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HLB Listener",
			fmt.Sprintf("Could not read HLB Listener ID %s: %v", state.ID.ValueString(), err),
		)
		return
	}

	// Overwrite items with refreshed state
	if listener.ALPNPolicy == "" {
		state.ALPNPolicy = types.StringNull()
	} else {
		state.ALPNPolicy = types.StringValue(listener.ALPNPolicy)
	}
	if listener.CertificateSecretsName == "" {
		state.CertificateSecretsName = types.StringNull()
	} else {
		state.CertificateSecretsName = types.StringValue(listener.CertificateSecretsName)
	}
	state.EnableDeletionProtection = types.BoolValue(listener.EnableDeletionProtection)
	state.OverprovisioningFactor = types.Float64Value(listener.OverprovisioningFactor)
	state.Port = types.Int64Value(int64(listener.Port))
	state.Protocol = types.StringValue(listener.Protocol)
	state.TargetGroupARN = types.StringValue(listener.TargetGroupARN)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *listenerAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan listenerAttachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state listenerAttachmentResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan
	input := &hlb.ListenerUpdate{}

	if !plan.ALPNPolicy.Equal(state.ALPNPolicy) {
		v := plan.ALPNPolicy.ValueString()
		input.ALPNPolicy = &v
	}

	if !plan.CertificateSecretsName.Equal(state.CertificateSecretsName) {
		v := plan.CertificateSecretsName.ValueString()
		input.CertificateSecretsName = &v
	}

	if !plan.EnableDeletionProtection.Equal(state.EnableDeletionProtection) {
		v := plan.EnableDeletionProtection.ValueBool()
		input.EnableDeletionProtection = &v
	}

	if !plan.OverprovisioningFactor.Equal(state.OverprovisioningFactor) {
		v := plan.OverprovisioningFactor.ValueFloat64()
		input.OverprovisioningFactor = &v
	}

	if !plan.Port.Equal(state.Port) {
		v := int(plan.Port.ValueInt64())
		input.Port = &v
	}

	if !plan.Protocol.Equal(state.Protocol) {
		v := plan.Protocol.ValueString()
		input.Protocol = &v
	}

	if !plan.TargetGroupARN.Equal(state.TargetGroupARN) {
		v := plan.TargetGroupARN.ValueString()
		input.TargetGroupARN = &v
	}

	// Update existing listener
	_, err := r.client.UpdateListener(ctx, state.LoadBalancerID.ValueString(), state.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating HLB Listener",
			fmt.Sprintf("Could not update listener %s: %v", state.ID.ValueString(), err),
		)
		return
	}

	// Fetch updated listener
	listener, err := r.client.GetListener(ctx, state.LoadBalancerID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HLB Listener",
			fmt.Sprintf("Could not read HLB Listener %s: %v", state.ID.ValueString(), err),
		)
		return
	}

	// Update resource state
	plan.ID = types.StringValue(listener.ID)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *listenerAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state listenerAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing listener
	err := r.client.DeleteListener(ctx, state.LoadBalancerID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting HLB Listener",
			fmt.Sprintf("Could not delete listener %s: %v", state.ID.ValueString(), err),
		)
		return
	}
}
