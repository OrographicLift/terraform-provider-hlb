# Create a new HLB load balancer
resource "hlb_load_balancer" "example" {
  name         = "example-hlb"
  internal     = false
  subnets      = ["subnet-12345678", "subnet-87654321"]
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
