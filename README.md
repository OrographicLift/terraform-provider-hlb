# HLB Terraform Provider

The HLB (Hero Load Balancer) Terraform Provider allows you to manage HLB resources in your Terraform configurations. This provider interacts with the HLB API to create, read, update, and delete Hosted Load Balancers and their associated resources in your AWS environment.

## Features

- Manage Hosted Load Balancers (HLBs) as Terraform resources
- Configure listeners for your HLBs
- Integrate HLB provisioning into your existing infrastructure-as-code workflows

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.13.x or later
- [Go](https://golang.org/doc/install) 1.23 or later (for building the provider plugin)
- AWS credentials configured (either through environment variables, shared credentials file, or IAM role)

## Building the Provider

Clone the repository:

```sh
git clone https://gitlab.guerraz.net/HLB/hlb-terraform-provider.git
cd hlb-terraform-provider
```

Build the provider using the provided Makefile:

```sh
make build
```

## Installing the Provider

To install the provider for local development and testing:

```sh
make install
```

This will build the provider and copy the binary to the appropriate local Terraform plugin directory.

## Using the Provider

To use the HLB provider in your Terraform configuration, include the following block:

```hcl
terraform {
  required_providers {
    hlb = {
      source  = "registry.terraform.io/guerraz.net/hlb"
      version = "1.0.0"
    }
  }
}

provider "hlb" {
  api_key     = var.hlb_api_key
  aws_region  = "us-west-2"  # Optional, will use AWS_REGION env var if not specified
  aws_profile = "my-profile" # Optional, will use default AWS authentication if not specified
}

resource "hlb_load_balancer" "example" {
  name     = "example-lb"
  internal = false
  subnets  = ["subnet-12345678", "subnet-87654321"]
  
  # Add other attributes as needed
}

resource "hlb_listener_attachment" "example" {
  load_balancer_id = hlb_load_balancer.example.id
  port             = 443
  protocol         = "HTTPS"
  target_group_arn = aws_lb_target_group.example.arn
  
  # Add other attributes as needed
}
```

Remember to replace `var.hlb_api_key` with your actual HLB API key.

## Documentation

For detailed documentation on the provider and its resources, please refer to the [HLB Provider Documentation](docs/index.md).

## Development

To set up your local development environment:

1. Clone the repository (as shown in the "Building the Provider" section)
2. Install dependencies: `go mod download`
3. Make your changes
4. Run tests: `make test`
5. Build and install the provider: `make install`

### Running Tests

- Run all tests: `make test`
- Run tests with race condition checking: `make test-race`
- Run only short tests: `make test-short`

### Code Quality

- Format your code: `make fmt`
- Run the linter: `make lint`

### Generating Documentation

To generate provider documentation:

```sh
make docs
```

## Contributing

Contributions to the HLB Terraform Provider are welcome! Please refer to the [contribution guidelines](CONTRIBUTING.md) for more information.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

For support, please contact your HLB service representative or open an issue on the GitLab repository.

