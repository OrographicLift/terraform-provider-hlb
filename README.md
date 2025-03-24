# HLB Terraform Provider

The HLB (Hero Load Balancer) Terraform Provider allows you to manage HLB resources in your Terraform configurations. This provider interacts with the HLB API to create, read, update, and delete Hosted Load Balancers and their associated resources in your AWS environment.

## Features

- Manage Hosted Load Balancers (HLBs) as Terraform resources
- Configure listeners for your HLBs
- Integrate HLB provisioning into your existing infrastructure-as-code workflows
- IAM based authentication

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) 0.13.x or later
- [Go](https://golang.org/doc/install) 1.24 or later (for building the provider plugin)
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
  partition   = "aws"        # Optional, defaults to "aws". Use "aws-dev" for development environment
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

The provider uses [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) to generate documentation from the provider schema and example files. The documentation is generated in the `docs/` directory and includes:

- Provider configuration and usage (`docs/index.md`)
- Resource documentation with schema details and examples (`docs/resources/*.md`)

The documentation is generated from:
- Schema descriptions in the provider code
- Example configurations in `examples/`
- Import examples in `examples/resources/*/import.sh`

To generate the documentation:

1. Ensure your code includes descriptive schema fields:
   ```go
   Schema: schema.Schema{
       Description: "Detailed description of the resource or field",
       Attributes: map[string]schema.Attribute{
           "field_name": schema.StringAttribute{
               Description: "Description of what this field does",
               Required:    true,
           },
       },
   }
   ```

2. Add example configurations in the `examples/` directory:
   - Provider configuration: `examples/provider/provider.tf`
   - Resource examples: `examples/resources/<resource_name>/resource.tf`
   - Import examples: `examples/resources/<resource_name>/import.sh`

3. Run the documentation generator using either:
   ```sh
   # Using make target
   make docs

   # Or directly with go generate
   go generate ./...

   # Or using the tool directly
   go tool tfplugindocs generate --provider-dir . --provider-name hlb
   ```

The generated documentation will be compatible with the Terraform Registry format and can be used both locally and when publishing the provider.

Note: Starting with Go 1.24, we use Go's built-in tool directive feature to manage the tfplugindocs dependency. This is configured in the go.mod file and eliminates the need for a separate tools directory.

### Release Process

Development of this provider happens on GitLab, but the Terraform Registry only supports GitHub for provider distribution. Therefore:

- Primary Development Repository: [GitLab](https://gitlab.guerraz.net/HLB/hlb-terraform-provider) - Submit contributions here
- Release Repository: [GitHub](https://github.com/OrographicLift/terraform-provider-hlb) - Mirror for Terraform Registry distribution

The provider uses GitHub Actions and GoReleaser to automate the release process. This automation ensures consistent builds across platforms and proper signing of releases for the Terraform Registry.

Note: The Terraform Registry requires the repository name to follow the pattern `terraform-provider-{NAME}` and be public on GitHub. The repository name must be lowercase.

#### Release Configuration

Two key files manage the release process:

1. `.github/workflows/release.yml`: GitHub Actions workflow that:
   - Triggers automatically on new tags (v*)
   - Sets up Go environment
   - Imports the GPG signing key
   - Runs GoReleaser to build and publish the release

2. `.goreleaser.yml`: Configures how releases are built and packaged:
   - Defines build targets (OS/architecture combinations)
   - Configures archive naming and format
   - Sets up checksums and GPG signing
   - Manages release artifacts

#### GPG Signing Key

Releases must be signed for the Terraform Registry. The project uses a dedicated GPG key stored as an organization secret:

- Location: GitHub Organization Secrets (https://github.com/organizations/OrographicLift/settings/secrets/actions/ZONEHERO_RELEASES_KEY)
- Key Requirements:
  - Email: info@zonehero.io
  - Key Length: ≥ 3072 bits
  - Validity: 14 months
  - No passphrase protection (for automated signing)
- Maintenance:
  - Key must be renewed every February
  - Both GitHub secret and Terraform Registry need the updated key

To generate a new signing key:

```bash
# Create key configuration
cat > key-config <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Name-Real: HLB Terraform Provider Releases
Name-Email: info@zonehero.io
Expire-Date: 14m
EOF

# Generate key
gpg --batch --generate-key key-config

# Export private key for GitHub Organization Secret, do not set a passphase
gpg --armor --export-secret-keys KEY_ID > zonehero_releases_key.asc

# Export public key for Terraform Registry
gpg --armor --export KEY_ID

# Once the key is updated in both GitHub and Terraform, delete the private key locally
gpg --delete-secret-keys KEY_ID
rm zonehero_releases_key.asc
```

#### Creating Releases

To create a new release:

1. Ensure all changes are committed and pushed
2. Create and push a new tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. The GitHub Action will automatically:
   - Build the provider for all supported platforms
   - Create a ZIP archive for each build
   - Generate checksums
   - Sign the release
   - Create a GitHub release
   - Upload all artifacts

The release will then be available for publishing to the Terraform Registry.

## Contributing

Contributions to the HLB Terraform Provider are welcome! Please submit all contributions to our [GitLab repository](https://gitlab.guerraz.net/HLB/hlb-terraform-provider). Refer to the [contribution guidelines](CONTRIBUTING.md) for more information.

Note: While there is a GitHub repository, it exists solely for Terraform Registry distribution. All development, issues, and merge requests should be submitted to GitLab.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

For support:
- Open an issue on our [GitLab repository](https://gitlab.guerraz.net/HLB/hlb-terraform-provider)
- Contact your HLB service representative
