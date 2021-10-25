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
  name    = aws_apigatewayv2_domain_name.flowsys.domain_name
  type    = "A"
  zone_id = data.aws_route53_zone.public.zone_id

  alias {
    name                   = aws_apigatewayv2_domain_name.flowsys.domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.flowsys.domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}