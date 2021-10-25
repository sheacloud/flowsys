resource "aws_cloudwatch_log_group" "firehose" {
  name = "/flowsys/parquet-flowlogs"
}

resource "aws_cloudwatch_log_stream" "firehose_destination" {
  name           = "DestinationDelivery"
  log_group_name = aws_cloudwatch_log_group.firehose.name
}

resource "aws_cloudwatch_log_stream" "firehose_backup" {
  name           = "BackupDelivery"
  log_group_name = aws_cloudwatch_log_group.firehose.name
}

resource "aws_cloudwatch_log_group" "ingestion_lambda" {
  name = "/aws/lambda/flowsys-ingestion"
}
