data "archive_file" "timestream_loader" {
  type        = "zip"
  source_file = "${path.module}/../services/timestream-loader/loader"
  output_path = "${path.module}/build/timestream-loader-function.zip"
}

resource "aws_lambda_function" "timestream_loader" {
  filename      = data.archive_file.timestream_loader.output_path
  function_name = "flowsys-timestream-loader"
  role          = aws_iam_role.timestream_loader_lambda_role.arn
  handler       = "loader"

  source_code_hash = data.archive_file.timestream_loader.output_base64sha256

  runtime = "go1.x"

  environment {
    variables = {
      "LOG_LEVEL"             = "TRACE"
      "TIMESTREAM_DB_NAME"    = aws_timestreamwrite_database.flowsys.database_name
      "TIMESTREAM_TABLE_NAME" = aws_timestreamwrite_table.flows.table_name
    }
  }

  lifecycle {
    ignore_changes = [last_modified]
  }
}

resource "aws_lambda_event_source_mapping" "timestream_loader_kinesis" {
  event_source_arn  = aws_kinesis_stream.flows.arn
  function_name     = aws_lambda_function.timestream_loader.arn
  starting_position = "LATEST"

  batch_size                         = 100
  maximum_batching_window_in_seconds = 300
  maximum_retry_attempts             = 3
  bisect_batch_on_function_error     = true
}
