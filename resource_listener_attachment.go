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
				Validators: []validator.String{
					stringvalidator.OneOf(
						"HTTP1Only",
						"HTTP2Only",
						"HTTP2Optional",
						"HTTP2Preferred",
						"None",
					),
				},
			},
			"certificate_secrets_name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the certificate secrets",
			},
			"enable_deletion_protection": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to enable deletion protection",
			},
			"load_balancer_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the load balancer",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"overprovisioning_factor": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(1.1),
				Description: "Overprovisioning factor for the listener",
				Validators: []validator.Float64{
					float64validator.AtLeast(1.0),
				},
			},
			"port": schema.Int64Attribute{
				Required:    true,
				Description: "Port on which the listener accepts connection",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "Protocol for connections from clients to the listener",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"HTTP",
						"HTTPS",
					),
				},
			},
			"target_group_arn": schema.StringAttribute{
				Required:    true,
				Description: "ARN of the target group",
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
		ALPNPolicy:               plan.ALPNPolicy.ValueString(),
		CertificateSecretsName:   plan.CertificateSecretsName.ValueString(),
		EnableDeletionProtection: plan.EnableDeletionProtection.ValueBool(),
		OverprovisioningFactor:   plan.OverprovisioningFactor.ValueFloat64(),
		Port:                     int(plan.Port.ValueInt64()),
		Protocol:                 plan.Protocol.ValueString(),
		TargetGroupARN:           plan.TargetGroupARN.ValueString(),
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
	state.ALPNPolicy = types.StringValue(listener.ALPNPolicy)
	state.CertificateSecretsName = types.StringValue(listener.CertificateSecretsName)
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
