resource "aws_kinesis_stream" "flows" {
  name             = "flowsys-flows"
  shard_count      = 1
  retention_period = 24

  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes",
  ]
}

// waiting for https://github.com/hashicorp/terraform-provider-aws/pull/20769 to be merged in

# resource "aws_kinesis_firehose_delivery_stream" "s3" {
#   name        = "flowsys-s3-stream"
#   destination = "extended_s3"

#   extended_s3_configuration {
#     bucket_arn = aws_s3_bucket.parquet_flowlogs.arn
#     role_arn   = aws_iam_role.firehose_role.arn

#     buffer_size = 128
#     buffer_interval = 300
#     error_output_prefix = "errors/"

#     cloudwatch_logging_options {
#       enabled         = true
#       log_group_name  = aws_cloudwatch_log_group.firehose.name
#       log_stream_name = aws_cloudwatch_log_stream.firehose_destination.name
#     }

#     processing_configuration {
#       enabled = true
#       processors {
#         type = "RecordDeAggregation"
#         parameters {
#           parameter_name = "SubRecordType"
#           parameter_value = "JSON"
#         }
#       }
#       processors {
#         type = "MetadataExtraction"
#         parameters {
#           parameter_name = "MetadataExtractionQuery"
#           parameter_value = "{date:.flow_start_milliseconds | (. / 1000 | strftime(\"%Y-%m-%d\"))}"
#         }
#         parameters {
#           parameter_name = "JsonParsingEngine"
#           parameter_value = "JQ-1.6"
#         }
#       }
#     }

#     data_format_conversion_configuration {
#       enabled = true
#       input_format_configuration {
#         deserializer {
#           open_x_json_ser_de {}
#         }
#       }

#       output_format_configuration {
#         serializer {
#           parquet_ser_de {}
#         }
#       }

#       schema_configuration {
#         role_arn = aws_iam_role.firehose_role.arn
#         version_id = "LATEST"
#         database_name = aws_glue_catalog_database.flowsys.name
#         table_name = aws_glue_catalog_table.flowlogs.name
#         region = data.aws_region.current.name
#       }
#     }
#   }

#   kinesis_source_configuration {
#     kinesis_stream_arn = aws_kinesis_stream.flows.arn
#     role_arn           = aws_iam_role.firehose_role.arn
#   }
# }
