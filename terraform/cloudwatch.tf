resource "aws_cloudwatch_log_group" "service" {
  name = "/ecs/flowsys/gateway"
}

resource "aws_cloudwatch_log_group" "firehose" {
  name = "/aws/kinesisfirehose/flowsys-s3-stream"
}

resource "aws_cloudwatch_log_stream" "firehose" {
  name           = "S3Delivery"
  log_group_name = aws_cloudwatch_log_group.firehose.name
}
