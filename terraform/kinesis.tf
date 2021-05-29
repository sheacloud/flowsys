resource "aws_kinesis_stream" "flows" {
  name             = "flowsys-flows-stream"
  shard_count      = 1
  retention_period = 24

  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes",
  ]
}

resource "aws_kinesis_stream_consumer" "analysis_service" {
  name       = "flowsys-analysis-service"
  stream_arn = aws_kinesis_stream.flows.arn
}


resource "aws_kinesis_firehose_delivery_stream" "s3" {
  name        = "flowsys-s3-stream"
  destination = "s3"

  s3_configuration {
    role_arn        = aws_iam_role.firehose_role.arn
    bucket_arn      = aws_s3_bucket.firehose.arn
    buffer_size     = 5
    buffer_interval = 300
    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = aws_cloudwatch_log_group.firehose.name
      log_stream_name = aws_cloudwatch_log_stream.firehose.name
    }
  }

  kinesis_source_configuration {
    kinesis_stream_arn = aws_kinesis_stream.flows.arn
    role_arn           = aws_iam_role.firehose_role.arn
  }
}
