// File: resource_load_balancer.go

package main

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"gitlab.guerraz.net/HLB/hlb-terraform-provider/hlb"
)

func resourceHLBLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHLBLoadBalancerCreate,
		ReadContext:   resourceHLBLoadBalancerRead,
		UpdateContext: resourceHLBLoadBalancerUpdate,
		DeleteContext: resourceHLBLoadBalancerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"name_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 6),
			},
			"internal": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"subnets": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				MinItems: 1,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"access_logs": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"launch_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"min_instance_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"max_instance_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"target_cpu_usage": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(10, 90),
						},
					},
				},
			},
			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_http2": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"idle_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntBetween(1, 4000),
			},
			"ip_address_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ipv4",
				ValidateFunc: validation.StringInSlice([]string{
					"ipv4", "dualstack", "dualstack-without-public-ipv4",
				}, false),
			},
			"preserve_host_header": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_cross_zone_load_balancing": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "avoid",
				ValidateFunc: validation.StringInSlice([]string{
					"full", "avoid", "off",
				}, false),
			},
			"client_keep_alive": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3600,
				ValidateFunc: validation.IntBetween(60, 604800),
			},
			"xff_header_processing_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "append",
				ValidateFunc: validation.StringInSlice([]string{
					"append", "preserve", "remove",
				}, false),
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			// Computed attributes
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceHLBLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	input := &hlb.LoadBalancerCreate{
		AccessLogs:                   expandAccessLogs(d.Get("access_logs").([]interface{})),
		ClientKeepAlive:              d.Get("client_keep_alive").(int),
		EnableCrossZoneLoadBalancing: d.Get("enable_cross_zone_load_balancing").(string),
		EnableDeletionProtection:     d.Get("enable_deletion_protection").(bool),
		EnableHttp2:                  d.Get("enable_http2").(bool),
		IdleTimeout:                  d.Get("idle_timeout").(int),
		Internal:                     d.Get("internal").(bool),
		IPAddressType:                d.Get("ip_address_type").(string),
		LaunchConfig:                 expandLaunchConfig(d.Get("launch_config").([]interface{})),
		Name:                         d.Get("name").(string),
		NamePrefix:                   d.Get("name_prefix").(string),
		PreserveHostHeader:           d.Get("preserve_host_header").(bool),
		SecurityGroups:               expandStringSet(d.Get("security_groups").(*schema.Set)),
		Subnets:                      expandStringSet(d.Get("subnets").(*schema.Set)),
		Tags:                         *expandTags(d.Get("tags").(map[string]interface{})),
		XffHeaderProcessingMode:      d.Get("xff_header_processing_mode").(string),
		ZoneID:                       d.Get("zone_id").(string),
		ZoneName:                     d.Get("zone_name").(string),
	}

	lb, err := client.CreateLoadBalancer(ctx, input)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(lb.ID)

	return resourceHLBLoadBalancerRead(ctx, d, meta)
}

func resourceHLBLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	lb, err := client.GetLoadBalancer(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if lb.AccessLogs != nil {
		d.Set("access_logs", []interface{}{flattenAccessLogs(lb.AccessLogs)})
	} else {
		d.Set("access_logs", nil)
	}
	d.Set("client_keep_alive", lb.ClientKeepAlive)
	d.Set("dns_name", lb.DNSName)
	d.Set("enable_cross_zone_load_balancing", lb.EnableCrossZoneLoadBalancing)
	d.Set("enable_deletion_protection", lb.EnableDeletionProtection)
	d.Set("enable_http2", lb.EnableHttp2)
	d.Set("idle_timeout", lb.IdleTimeout)
	d.Set("internal", lb.Internal)
	d.Set("ip_address_type", lb.IPAddressType)
	if lb.LaunchConfig != nil {
		d.Set("launch_config", []interface{}{flattenLaunchConfig(lb.LaunchConfig)})
	} else {
		d.Set("launch_config", nil)
	}
	d.Set("name", lb.Name)
	d.Set("preserve_host_header", lb.PreserveHostHeader)
	d.Set("security_groups", lb.SecurityGroups)
	d.Set("state", lb.State)
	d.Set("subnets", lb.Subnets)
	d.Set("tags", flattenTags(lb.Tags))
	d.Set("xff_header_processing_mode", lb.XffHeaderProcessingMode)
	d.Set("zone_id", lb.ZoneID)
	d.Set("zone_name", lb.ZoneName)

	return nil
}

func resourceHLBLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	input := &hlb.LoadBalancerUpdate{
		AccessLogs:     expandAccessLogs(d.Get("access_logs").([]interface{})),
		SecurityGroups: expandStringSet(d.Get("security_groups").(*schema.Set)),
		Tags:           expandTags(d.Get("tags").(map[string]interface{})),
	}

	if d.HasChange("name") {
		v := d.Get("name").(string)
		input.Name = &v
	}

	if d.HasChange("enable_deletion_protection") {
		v := d.Get("enable_deletion_protection").(bool)
		input.EnableDeletionProtection = &v
	}

	if d.HasChange("enable_http2") {
		v := d.Get("enable_http2").(bool)
		input.EnableHttp2 = &v
	}

	if d.HasChange("idle_timeout") {
		v := d.Get("idle_timeout").(int)
		input.IdleTimeout = &v
	}

	if d.HasChange("preserve_host_header") {
		v := d.Get("preserve_host_header").(bool)
		input.PreserveHostHeader = &v
	}

	if d.HasChange("enable_cross_zone_load_balancing") {
		v := d.Get("enable_cross_zone_load_balancing").(string)
		input.EnableCrossZoneLoadBalancing = &v
	}

	if d.HasChange("client_keep_alive") {
		v := d.Get("client_keep_alive").(int)
		input.ClientKeepAlive = &v
	}

	if d.HasChange("xff_header_processing_mode") {
		v := d.Get("xff_header_processing_mode").(string)
		input.XffHeaderProcessingMode = &v
	}

	if d.HasChange("launch_config") {
		input.LaunchConfig = expandLaunchConfig(d.Get("launch_config").([]interface{}))
	}

	_, err := client.UpdateLoadBalancer(ctx, d.Id(), input)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceHLBLoadBalancerRead(ctx, d, meta)
}

func resourceHLBLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*hlb.Client)

	err := client.DeleteLoadBalancer(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func expandStringSet(set *schema.Set) []string {
	result := make([]string, set.Len())
	for i, v := range set.List() {
		result[i] = v.(string)
	}
	return result
}

func expandAccessLogs(l []interface{}) *hlb.AccessLogs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	accessLogs := &hlb.AccessLogs{
		Enabled: m["enabled"].(bool),
		Bucket:  m["bucket"].(string),
		Prefix:  m["prefix"].(string),
	}

	return accessLogs
}

func flattenAccessLogs(accessLogs *hlb.AccessLogs) map[string]interface{} {
	if accessLogs == nil {
		return nil
	}

	return map[string]interface{}{
		"enabled": accessLogs.Enabled,
		"bucket":  accessLogs.Bucket,
		"prefix":  accessLogs.Prefix,
	}
}

func expandLaunchConfig(l []interface{}) *hlb.LaunchConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	launchConfig := &hlb.LaunchConfig{
		InstanceType:     m["instance_type"].(string),
		MinInstanceCount: m["min_instance_count"].(int),
		MaxInstanceCount: m["max_instance_count"].(int),
		TargetCPUUsage:   m["target_cpu_usage"].(int),
	}

	return launchConfig
}

func flattenLaunchConfig(launchConfig *hlb.LaunchConfig) map[string]interface{} {
	if launchConfig == nil {
		return nil
	}

	return map[string]interface{}{
		"instance_type":      launchConfig.InstanceType,
		"min_instance_count": launchConfig.MinInstanceCount,
		"max_instance_count": launchConfig.MaxInstanceCount,
		"target_cpu_usage":   launchConfig.TargetCPUUsage,
	}
}

func expandTags(m map[string]interface{}) *map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v.(string)
	}
	return &result
}

func flattenTags(tags map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(tags))
	for k, v := range tags {
		result[k] = v
	}
	return result
}
