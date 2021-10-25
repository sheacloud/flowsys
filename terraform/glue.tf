resource "aws_glue_catalog_database" "flowsys" {
  name = "flowsys"
}

resource "aws_glue_catalog_table" "flowlogs" {
  name          = "flowlogs"
  database_name = aws_glue_catalog_database.flowsys.name
  table_type    = "EXTERNAL_TABLE"
  parameters = {
    EXTERNAL                        = "TRUE"
    "parquet.compression"           = "SNAPPY"
    "projection.enabled"            = "true"
    "projection.flow_start_date.format" = "yyyy-MM-dd"
    "projection.flow_start_date.range"  = "NOW-3YEARS,NOW"
    "projection.flow_start_date.type"   = "date"
  }

  storage_descriptor {
    location      = "s3://${aws_s3_bucket.parquet_flowlogs.bucket}/"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"
      parameters = {
        "serialization.format" = "1"
      }
    }

    columns {
      name    = "source_ipv4_address"
      type    = "string"
      comment = ""
    }
    columns {
      name    = "destination_ipv4_address"
      type    = "string"
      comment = ""
    }
    columns {
      name    = "source_ipv6_address"
      type    = "string"
      comment = ""
    }
    columns {
      name    = "destination_ipv6_address"
      type    = "string"
      comment = ""
    }
    columns {
      name    = "source_port"
      type    = "int"
      comment = ""
    }
    columns {
      name    = "destination_port"
      type    = "int"
      comment = ""
    }
    columns {
      name    = "ip_protocol_version"
      type    = "tinyint"
      comment = ""
    }
    columns {
      name    = "transport_protocol"
      type    = "smallint"
      comment = ""
    }
    columns {
      name    = "flow_start_milliseconds"
      type    = "bigint"
      comment = ""
    }
    columns {
      name    = "flow_end_milliseconds"
      type    = "bigint"
      comment = ""
    }
    columns {
      name    = "flow_octet_count"
      type    = "bigint"
      comment = ""
    }
    columns {
      name    = "flow_packet_count"
      type    = "bigint"
      comment = ""
    }
    columns {
      name    = "reverse_flow_octet_count"
      type    = "bigint"
      comment = ""
    }
    columns {
      name    = "reverse_flow_packet_count"
      type    = "bigint"
      comment = ""
    }
  }

  partition_keys {
    name = "flow_start_date"
    type = "date"
  }
}