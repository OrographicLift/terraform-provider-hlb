package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

func resourceHLBListenerAttachment() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHLBListenerAttachmentCreate,
		ReadContext:   resourceHLBListenerAttachmentRead,
		UpdateContext: resourceHLBListenerAttachmentUpdate,
		DeleteContext: resourceHLBListenerAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"alpn_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"HTTP1Only", "HTTP2Only", "HTTP2Optional", "HTTP2Preferred", "None",
				}, false),
			},
			"certificate_secrets_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"overprovisioning_factor": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  1.1,
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"protocol": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"HTTP", "HTTPS",
				}, false),
			},
			"target_group_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceHLBListenerAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	input := &hlb.ListenerCreate{
		ALPNPolicy:               d.Get("alpn_policy").(string),
		CertificateSecretsName:   d.Get("certificate_secrets_name").(string),
		EnableDeletionProtection: d.Get("enable_deletion_protection").(bool),
		OverprovisioningFactor:   d.Get("overprovisioning_factor").(float64),
		Port:                     d.Get("port").(int),
		Protocol:                 d.Get("protocol").(string),
		TargetGroupARN:           d.Get("target_group_arn").(string),
	}

	loadBalancerID := d.Get("load_balancer_id").(string)
	listener, err := client.CreateListener(ctx, loadBalancerID, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(listener.ID)
	return resourceHLBListenerAttachmentRead(ctx, d, meta)
}

func resourceHLBListenerAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	loadBalancerID := d.Get("load_balancer_id").(string)
	listener, err := client.GetListener(ctx, loadBalancerID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("alpn_policy", listener.ALPNPolicy)
	d.Set("certificate_secrets_name", listener.CertificateSecretsName)
	d.Set("enable_deletion_protection", listener.EnableDeletionProtection)
	d.Set("overprovisioning_factor", listener.OverprovisioningFactor)
	d.Set("port", listener.Port)
	d.Set("protocol", listener.Protocol)
	d.Set("target_group_arn", listener.TargetGroupARN)

	return nil
}

func resourceHLBListenerAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	input := &hlb.ListenerUpdate{}

	if d.HasChange("alpn_policy") {
		input.ALPNPolicy = new(string)
		*input.ALPNPolicy = d.Get("alpn_policy").(string)
	}
	if d.HasChange("certificate_secrets_name") {
		input.CertificateSecretsName = new(string)
		*input.CertificateSecretsName = d.Get("certificate_secrets_name").(string)
	}
	if d.HasChange("enable_deletion_protection") {
		v := d.Get("enable_deletion_protection").(bool)
		input.EnableDeletionProtection = &v
	}
	if d.HasChange("overprovisioning_factor") {
		v := d.Get("overprovisioning_factor").(float64)
		input.OverprovisioningFactor = &v
	}
	if d.HasChange("port") {
		input.Port = new(int)
		*input.Port = d.Get("port").(int)
	}
	if d.HasChange("protocol") {
		input.Protocol = new(string)
		*input.Protocol = d.Get("protocol").(string)
	}
	if d.HasChange("target_group_arn") {
		input.TargetGroupARN = new(string)
		*input.TargetGroupARN = d.Get("target_group_arn").(string)
	}

	loadBalancerID := d.Get("load_balancer_id").(string)
	_, err := client.UpdateListener(ctx, loadBalancerID, d.Id(), input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHLBListenerAttachmentRead(ctx, d, meta)
}

func resourceHLBListenerAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	loadBalancerID := d.Get("load_balancer_id").(string)
	err := client.DeleteListener(ctx, loadBalancerID, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
