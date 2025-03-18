# Create a new HLB listener attachment
resource "hlb_listener_attachment" "front_end" {
  load_balancer_id        = hlb_load_balancer.example.id
  port                    = 443
  protocol                = "HTTPS"
  target_group_arn        = aws_lb_target_group.front_end.arn
  certificate_secrets_arn = "arn:aws:secretsmanager:us-west-2:123456789012:secret:my-certificate-secret-123abc"
  alpn_policy             = "HTTP2Preferred"
  overprovisioning_factor = 1.1 # Optional, defaults to 1.1
}
