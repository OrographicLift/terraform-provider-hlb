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
			"load_balancer_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"certificate_secrets_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"alpn_policy": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"HTTP1Only", "HTTP2Only", "HTTP2Optional", "HTTP2Preferred", "None",
				}, false),
			},
		},
	}
}

func resourceHLBListenerAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	input := &hlb.CreateListenerInput{
		Port:                  d.Get("port").(int),
		Protocol:              d.Get("protocol").(string),
		TargetGroupARN:        d.Get("target_group_arn").(string),
		CertificateSecretsARN: d.Get("certificate_secrets_arn").(string),
		ALPNPolicy:            d.Get("alpn_policy").(string),
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

	d.Set("port", listener.Port)
	d.Set("protocol", listener.Protocol)
	d.Set("target_group_arn", listener.TargetGroupARN)
	d.Set("certificate_secrets_arn", listener.CertificateSecretsARN)
	d.Set("alpn_policy", listener.ALPNPolicy)

	return nil
}

func resourceHLBListenerAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	input := &hlb.UpdateListenerInput{}

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
	if d.HasChange("certificate_secrets_arn") {
		input.CertificateSecretsARN = new(string)
		*input.CertificateSecretsARN = d.Get("certificate_secrets_arn").(string)
	}
	if d.HasChange("alpn_policy") {
		input.ALPNPolicy = new(string)
		*input.ALPNPolicy = d.Get("alpn_policy").(string)
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