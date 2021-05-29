resource "aws_cloudwatch_log_group" "service" {
  name = "/ecs/flowsys/${var.service_name}"
}
