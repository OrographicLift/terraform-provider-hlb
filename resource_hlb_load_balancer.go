// File: resource_hlb_load_balancer.go
package main

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceHLBLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHLBLoadBalancerCreate,
		ReadContext:   resourceHLBLoadBalancerRead,
		UpdateContext: resourceHLBLoadBalancerUpdate,
		DeleteContext: resourceHLBLoadBalancerDelete,

		Schema: map[string]*schema.Schema{
			"target_group_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^arn:aws:elasticloadbalancing:.*$`), "must be a valid target group ARN"),
			},
			"fqdn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$`), "must be a valid FQDN"),
			},
			"route53_zone_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^Z[A-Z0-9]{10,}$`), "must be a valid Route53 Zone ID"),
			},
			"config": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHLBLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	hlb := &HLB{
		TargetGroupARN: d.Get("target_group_arn").(string),
		FQDN:           d.Get("fqdn").(string),
		Route53ZoneID:  d.Get("route53_zone_id").(string),
		Config:         d.Get("config").(map[string]interface{}),
	}

	createdHLB, err := client.CreateHLB(hlb)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(createdHLB.ID)

	return resourceHLBLoadBalancerRead(ctx, d, m)
}

func resourceHLBLoadBalancerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	hlb, err := client.GetHLB(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("target_group_arn", hlb.TargetGroupARN)
	d.Set("fqdn", hlb.FQDN)
	d.Set("route53_zone_id", hlb.Route53ZoneID)
	d.Set("config", hlb.Config)
	d.Set("status", hlb.Status)
	d.Set("created_at", hlb.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", hlb.UpdatedAt.Format(time.RFC3339))

	return nil
}

func resourceHLBLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	hlb := &HLB{
		ID:             d.Id(),
		TargetGroupARN: d.Get("target_group_arn").(string),
		FQDN:           d.Get("fqdn").(string),
		Route53ZoneID:  d.Get("route53_zone_id").(string),
		Config:         d.Get("config").(map[string]interface{}),
	}

	_, err := client.UpdateHLB(d.Id(), hlb)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHLBLoadBalancerRead(ctx, d, m)
}

func resourceHLBLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	err := client.DeleteHLB(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
