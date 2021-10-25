resource "aws_s3_bucket" "parquet_flowlogs" {
  bucket = "${var.bucket_prefix}-parquet-flowlogs"
  acl    = "private"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}
