data "aws_route53_zone" "public" {
  name         = var.public_domain_name
  private_zone = false
}

resource "aws_route53_record" "flowsys_validation" {
  for_each = {
    for dvo in aws_acm_certificate.flowsys.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.public.zone_id
}

resource "aws_route53_record" "flowsys" {
  zone_id = data.aws_route53_zone.public.zone_id
  name    = "flowsys.${var.public_domain_name}"
  type    = "A"

  alias {
    name                   = aws_lb.flowsys.dns_name
    zone_id                = aws_lb.flowsys.zone_id
    evaluate_target_health = true
  }
}
