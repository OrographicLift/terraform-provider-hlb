# Configure the HLB Provider
terraform {
  required_providers {
    hlb = {
      source  = "registry.terraform.io/OrographicLift/hlb"
      version = "~> 1.0"
    }
  }
}

provider "hlb" {
  api_key     = "your-api-key" # Can also be set with HLB_API_KEY environment variable
  aws_region  = "us-west-2"    # Optional, will use AWS_REGION env var if not specified
  aws_profile = "my-profile"   # Optional, will use AWS_PROFILE if not specified
  partition   = "aws"          # Optional, defaults to "aws". Use "aws-dev" if you have been granted access to the development environment
}
