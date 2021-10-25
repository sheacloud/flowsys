resource "aws_apigatewayv2_api" "flowsys" {
  name          = "flowsys"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id = aws_apigatewayv2_api.flowsys.id
  name   = "$default"
  auto_deploy = true
}

resource "aws_apigatewayv2_integration" "flowsys_ingestion" {
  api_id           = aws_apigatewayv2_api.flowsys.id
  integration_type = "AWS_PROXY"

  connection_type           = "INTERNET"
  description               = "flowsys ingestion lambda"
  integration_method        = "POST"
  integration_uri           = aws_lambda_function.flowsys_ingestion.invoke_arn
  payload_format_version = "2.0"

  depends_on = [
    aws_lambda_permission.api_gateway,
  ]
}

resource "aws_apigatewayv2_route" "flowsys" {
  api_id    = aws_apigatewayv2_api.flowsys.id
  route_key = "$default"

  target = "integrations/${aws_apigatewayv2_integration.flowsys_ingestion.id}"
}

resource "aws_apigatewayv2_domain_name" "flowsys" {
  domain_name = "flowsys.${var.public_domain_name}"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.flowsys.arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}

resource "aws_apigatewayv2_api_mapping" "flowsys" {
  api_id      = aws_apigatewayv2_api.flowsys.id
  domain_name = aws_apigatewayv2_domain_name.flowsys.id
  stage       = aws_apigatewayv2_stage.default.id
}