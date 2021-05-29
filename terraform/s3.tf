resource "aws_s3_bucket" "firehose" {
  bucket = var.flowsys_firehose_bucket_name
  acl    = "private"

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm = "AES256"
      }
    }
  }
}
