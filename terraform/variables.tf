variable "vpc_id" {
  type = string
}

variable "application_port" {
  type    = number
  default = 8080
}

variable "envoy_image" {
  type    = string
  default = "840364872350.dkr.ecr.us-east-1.amazonaws.com/aws-appmesh-envoy:v1.17.3.0-prod"
}

variable "alb_ingress_cidrs" {
  type = list(string)
}

variable "public_domain_name" {
  type = string
}

variable "flowsys_firehose_bucket_name" {
  type = string
}
