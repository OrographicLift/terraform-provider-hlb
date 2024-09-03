# Terraform Provider for Hosted Load Balancer (HLB)

## Overview

The Terraform Provider for Hosted Load Balancer (HLB) allows you to manage HLB resources in your Terraform configurations. This provider interacts with the HLB Provisioner API to create, read, update, and delete Hosted Load Balancers in your AWS environment.

## Features

- Manage Hosted Load Balancers (HLBs) as Terraform resources
- Integrate HLB provisioning into your existing infrastructure-as-code workflows
- Supports creating, updating, and deleting HLBs

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.13.x or later
- [Go](https://golang.org/doc/install) 1.23 or later (for building the provider plugin)

## Building the Provider

1. Clone the repository
```sh
git clone https://github.com/your-org/terraform-provider-hlb.git
```

2. Enter the provider directory
```sh
cd terraform-provider-hlb
```

3. Build the provider
```sh
make build
```

## Installing the Provider

After building the provider, you can install it locally using the following command:

```sh
make install
```

This will install the provider in your local Terraform plugin directory.

## Using the Provider

To use the HLB provider in your Terraform configuration, include the following block:

```hcl
terraform {
  required_providers {
    hlb = {
      source = "your-org/hlb"
      version = "1.0.0"  # Use the appropriate version
    }
  }
}

provider "hlb" {
  api_key = "your-api-key"
  api_url = "https://api.hlb.example.com"
}

resource "hlb_load_balancer" "example" {
  target_group_arn = "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/example-tg/1234567890abcdef"
  fqdn             = "example.yourdomain.com"
  route53_zone_id  = "Z1234567890ABCDEF"
  
  config = {
    "key1" = "value1"
    "key2" = "value2"
  }
}
```

Remember to replace `"your-api-key"` and `"https://api.hlb.example.com"` with your actual HLB API key and API URL.

## Documentation

For detailed documentation on the provider and its resources, please refer to the [HLB Provider Documentation](https://docs.example.com/hlb-provider).

## Contributing

Contributions to the HLB Terraform Provider are welcome! Please refer to the [contribution guidelines](CONTRIBUTING.md) for more information.

## License

This Terraform provider is released under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

For support, please contact your HLB service representative or open an issue on the GitHub repository.n