resource "aws_iam_role" "ingestion_lambda" {
  name = "flowsys-ingestion-lambda-role"

  assume_role_policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_policy" "ingestion_lambda" {
  name = "flowsys-ingestion-lambda-policy"
  path = "/"
  description = "Allow lambda function to write to Kinesis"

  policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "kinesis:PutRecord",
          "kinesis:PutRecords"
        ]
        "Resource": aws_kinesis_stream.flows.arn
      },
      {
        "Effect": "Allow",
        "Action": [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        "Resource": "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/flowsys-ingestion:*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ingestion_lambda" {
  role       = aws_iam_role.ingestion_lambda.name
  policy_arn = aws_iam_policy.ingestion_lambda.arn
}

resource "aws_iam_role" "firehose_role" {
  name = "flowsys-kinesis-firehose-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "firehose.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_policy" "firehose_policy" {
  name        = "flowsys-kinesis-firehose-policy"
  path        = "/"
  description = "Allow Kinesis Firehose to pull from Kinesis and write to S3 via a glue table catalog"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "kinesis:DescribeStream",
          "kinesis:GetShardIterator",
          "kinesis:GetRecords",
          "kinesis:ListShards"
        ]
        Effect   = "Allow"
        Resource = aws_kinesis_stream.flows.arn
      },
      {
        Action = [
          "logs:PutLogEvents",
        ]
        Effect   = "Allow"
        Resource = aws_cloudwatch_log_stream.firehose_destination.arn
      },
      {
        Action = [
          "s3:AbortMultipartUpload",
          "s3:GetBucketLocation",
          "s3:GetObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:PutObject"
        ]
        Effect = "Allow"
        Resource = [
          aws_s3_bucket.parquet_flowlogs.arn,
          "${aws_s3_bucket.parquet_flowlogs.arn}/*"
        ]
      },
      {
        Action = [
          "glue:GetTable",
          "glue:GetTableVersion",
          "glue:GetTableVersions"
        ],
        Effect = "Allow",
        Resource = [
          "arn:aws:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:catalog",
          aws_glue_catalog_database.flowsys.arn,
          aws_glue_catalog_table.flowlogs.arn
        ]
      }
    ]
  })
}


resource "aws_iam_role_policy_attachment" "firehose" {
  role       = aws_iam_role.firehose_role.name
  policy_arn = aws_iam_policy.firehose_policy.arn
}