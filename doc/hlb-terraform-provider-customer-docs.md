# Hero Load Balancer (HLB) Terraform Provider

The Hero Load Balancer (HLB) Terraform Provider allows you to manage HLB resources in your Terraform configurations. This document covers the available resources and their usage.

## Provider Configuration

To use the HLB provider, you need to configure it in your Terraform configuration. Here's a basic example:

```hcl
terraform {
  required_providers {
    hlb = {
      source = "orographiclift/hlb"
      version = "~> 1.0"
    }
  }
}

provider "hlb" {
  api_key     = "your-api-key"
  aws_region  = "us-west-2"    # Optional, will use AWS_REGION env var if not specified
  aws_profile = "my-profile"   # Optional, will use default AWS authentication if not specified
  partition   = "aws"          # Optional, defaults to "aws". Use "aws-dev" for development environment
}
```

### Provider Arguments

The following arguments are supported:

* `api_key` - (Required) Your HLB API key
* `aws_region` - (Optional) AWS region to use. If not specified, will use AWS_REGION environment variable
* `aws_profile` - (Optional) AWS profile to use. If not specified, will use default AWS authentication
* `partition` - (Optional) Partition to use. Valid values are "aws" (default) for production and "aws-dev" for development environment

## Resources

### hlb_load_balancer

The `hlb_load_balancer` resource allows you to create and manage a Hero Load Balancer.

#### Example Usage

```hcl
resource "hlb_load_balancer" "example" {
  name        = "example-hlb"
  internal    = false
  subnets     = ["subnet-12345678", "subnet-87654321"]
  enable_http2 = true

  access_logs {
    bucket  = "my-logs-bucket"
    prefix  = "my-hlb-logs"
    enabled = true
  }

  # Optional: customize the launch configuration
  launch_config {
    instance_type      = "c7g.large"
    min_instance_count = 2
    target_cpu_usage   = 40
  }

  tags = {
    Environment = "production"
  }
}

output "load_balancer_dns" {
  value = hlb_load_balancer.example.dns_name
}
```

#### Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the HLB. Must be unique within your account, max 32 characters, alphanumeric and hyphens only.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `internal` - (Optional) If true, the HLB will be internal. Defaults to false.
* `subnets` - (Optional) List of subnet IDs to attach to the HLB.
* `security_groups` - (Optional) List of security group IDs to assign to the HLB.
* `access_logs` - (Optional) An Access Logs block. Details below.
* `launch_config` - (Optional) A Launch Configuration block. Details below.
* `enable_deletion_protection` - (Optional) Prevents deletion via the API. Defaults to false.
* `enable_http2` - (Optional) Enables HTTP/2. Defaults to true.
* `idle_timeout` - (Optional) Idle timeout in seconds. Default: 60.
* `ip_address_type` - (Optional) IP address type. Valid values: `ipv4`, `dualstack`, `dualstack-without-public-ipv4`.
* `preserve_host_header` - (Optional) Preserves the Host header. Defaults to false.
* `enable_cross_zone_load_balancing` - (Optional) Cross-zone load balancing. Values: "full", "avoid" (default), "off".
* `client_keep_alive` - (Optional) Client keep alive in seconds (60-604800). Default: 3600.
* `xff_header_processing_mode` - (Optional) X-Forwarded-For header processing. Values: `append`, `preserve`, `remove`. Default: `append`.
* `tags` - (Optional) A map of tags to assign to the resource.

##### Access Logs Arguments

For `access_logs` the following attributes are supported:

* `bucket` - (Required) The S3 bucket name to store the logs in.
* `prefix` - (Optional) The S3 bucket prefix. Logs are stored in the root if not configured.
* `enabled` - (Optional) Boolean to enable / disable access_logs. Defaults to false.

##### Launch Config Arguments

For `launch_config` the following attributes are supported:

* `instance_type` - (Optional) The EC2 instance type to use for the load balancer nodes. If not specified, the backend will choose an appropriate default (currently c7g.medium).
* `min_instance_count` - (Optional) The minimum number of instances to maintain (integer, minimum 1). This will be rounded up to a multiple of the number of AZ. If not specified, the backend will choose an appropriate default.
* `max_instance_count` - (Optional) The maximum number of instances to maintain (integer, minimum 1). This will be rounded down to a multiple of the number of AZ. If not specified, the backend will choose 10 x `min_instance_count`.
* `target_cpu_usage` - (Optional) The target CPU usage percentage (10-90) for auto-scaling. If not specified, the backend will choose an appropriate default.

Note: The launch configuration is entirely optional. If not specified, the backend will use appropriate defaults for all fields. This allows the backend to automatically adjust defaults for all customers who haven't explicitly set these values.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the HLB.
* `uri` - The full URI of the HLB.
* `dns_name` - The DNS name of the HLB (provided by the API).
* `zone_id` - The canonical hosted zone ID of the HLB.

### hlb_listener_attachment

The `hlb_listener_attachment` resource allows you to create and manage a listener for your Hero Load Balancer.

#### Example Usage

```hcl
resource "hlb_listener_attachment" "front_end" {
  load_balancer_id     = hlb_load_balancer.front_end.id
  port                 = 443
  protocol             = "HTTPS"
  target_group_arn     = aws_lb_target_group.front_end.arn
  certificate_secrets_arn = "arn:aws:secretsmanager:us-west-2:123456789012:secret:my-certificate-secret-123abc"
  alpn_policy          = "HTTP2Preferred"
}
```

#### Argument Reference

The following arguments are supported:

* `load_balancer_id` - (Required) The ID of the HLB.
* `port` - (Required) The port on which the load balancer is listening.
* `target_group_arn` - (Required) The ARN of the target group to which to route traffic.
* `overprovisioning_factor` - (Optional) When the load balancer `enable_cross_zone_load_balancing` is set to `avoid` (default), allows the load balancer to send at most '`overprovisioning_factor` * the average amount of request' to any single instance. Applies to the target group, behaviour is undefined if different values are assigned to different listeners with the same target group. Defaults to `1.1`
* `protocol` - (Optional) The protocol for connections from clients to the load balancer. Valid values are `HTTP` and `HTTPS`. Defaults to `HTTP`.
* `alpn_policy` - (Optional) The Application-Layer Protocol Negotiation (ALPN) policy. Valid values are `HTTP1Only`, `HTTP2Only`, `HTTP2Optional`, `HTTP2Preferred`, and `None`.
* `certificate_secrets_arn` - (Optional) ARN of the secret in AWS Secrets Manager containing the SSL server certificate. Required if the protocol is HTTPS.

#### Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the listener.
* `uri` - The full URI of the listener.

## Import

HLB resources can be imported using the `id`, e.g.,

```
$ terraform import hlb_load_balancer.test hlb-1234567890abcdef
$ terraform import hlb_listener_attachment.test lis-1234567890abcdef
```

For more information on using the HLB Terraform Provider, please refer to our full documentation or contact our support team.
