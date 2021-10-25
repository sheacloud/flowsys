data "archive_file" "flowsys_ingestion" {
  type        = "zip"
  source_file = "${path.module}/../build/ingestion"
  output_path = "${path.module}/build/ingestion.zip"
}

resource "aws_lambda_function" "flowsys_ingestion" {
  filename      = data.archive_file.flowsys_ingestion.output_path
  function_name = "flowsys-ingestion"
  role          = aws_iam_role.ingestion_lambda.arn
  handler       = "ingestion"

  timeout = 30

  source_code_hash = data.archive_file.flowsys_ingestion.output_base64sha256

  runtime = "go1.x"

  environment {
    variables = {
      "LOG_LEVEL"             = "INFO"
      "KINESIS_STREAM_NAME" = aws_kinesis_stream.flows.name
    }
  }

  lifecycle {
    ignore_changes = [last_modified]
  }
}

resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "allow-flowsys-apigateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.flowsys_ingestion.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.flowsys.execution_arn}/*/*"
}