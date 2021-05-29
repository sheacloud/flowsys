resource "aws_acm_certificate" "flowsys" {
  domain_name       = "flowsys.${var.public_domain_name}"
  validation_method = "DNS"
}

resource "aws_acm_certificate_validation" "flowsys" {
  certificate_arn         = aws_acm_certificate.flowsys.arn
  validation_record_fqdns = [for record in aws_route53_record.flowsys_validation : record.fqdn]
}
