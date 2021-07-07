resource "aws_iam_policy" "gateway_ecs_task_policy" {
  name        = "flowsys-gateway-task-policy"
  path        = "/"
  description = "Allow envoy proxy to write to cloudwatch logs"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:*",
          "s3:PutObject",
          "appmesh:*",
          "cloudwatch:*",
          "xray:*"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}


resource "aws_iam_role" "gateway_ecs_task_role" {
  name = "flowsys-gateway-ecs-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "gateway_ecs_task_role_custom_attach" {
  role       = aws_iam_role.gateway_ecs_task_role.name
  policy_arn = aws_iam_policy.gateway_ecs_task_policy.arn
}


resource "aws_iam_role" "gateway_ecs_task_execution_role" {
  name = "flowsys-gateway-ecs-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "gateway_ecs_task_execution_role_aws_attach" {
  role       = aws_iam_role.gateway_ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
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
  description = "Allow Kinesis Firehose to pull from Kinesis and write to S3"

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
        Resource = aws_cloudwatch_log_stream.firehose.arn
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
          aws_s3_bucket.firehose.arn,
          "${aws_s3_bucket.firehose.arn}/*"
        ]
      }
    ]
  })
}


resource "aws_iam_role_policy_attachment" "firehose" {
  role       = aws_iam_role.firehose_role.name
  policy_arn = aws_iam_policy.firehose_policy.arn
}




resource "aws_iam_role" "timestream_loader_lambda_role" {
  name = "flowsys-timestream-loader-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_policy" "timestream_loader_lambda_policy" {
  name        = "flowsys-timestream-loader-lambda-policy"
  path        = "/"
  description = "Allow Timestream Loader Lambda function to read from kinesis data stream and write to Timestream"

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
          "kinesis:ListStreams",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action = [
          "logs:PutLogEvents",
          "logs:CreateLogStream"
        ]
        Effect   = "Allow"
        Resource = "${aws_cloudwatch_log_group.timestream_loader_lambda.arn}:*"
      },
      {
        Action = [
          "timestream:WriteRecords"
        ]
        Effect   = "Allow"
        Resource = aws_timestreamwrite_table.flows.arn
      },
      {
        Action = [
          "timestream:DescribeEndpoints"
        ]
        Effect   = "Allow"
        Resource = "*"
      }
    ]
  })
}


resource "aws_iam_role_policy_attachment" "timestream_loader_lambda" {
  role       = aws_iam_role.timestream_loader_lambda_role.name
  policy_arn = aws_iam_policy.timestream_loader_lambda_policy.arn
}
