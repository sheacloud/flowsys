data "template_file" "task_definition" {
  template = file("${path.module}/../../task-definitions/service.json")
  vars = {
    IMAGE                    = aws_ecr_repository.repo.repository_url
    APPMESH_VIRTUAL_NODE_ARN = "mesh/${var.appmesh_name}/virtualNode/${aws_appmesh_virtual_node.service.name}"
    SERVICE_NAME             = var.service_name
    APPLICATION_PORT         = var.application_port
    ENVOY_IMAGE              = var.envoy_image
  }
}


resource "aws_ecs_task_definition" "task_definition" {
  family                   = "flowsys-${var.service_name}"
  container_definitions    = data.template_file.task_definition.rendered
  requires_compatibilities = ["FARGATE"]

  network_mode = "awsvpc"

  cpu                = 256
  memory             = 512
  task_role_arn      = aws_iam_role.task_role.arn
  execution_role_arn = aws_iam_role.task_execution_role.arn

  proxy_configuration {
    type           = "APPMESH"
    container_name = "envoy"
    properties = {
      AppPorts         = "8080"
      EgressIgnoredIPs = "169.254.170.2,169.254.169.254"
      IgnoredUID       = "1337"
      ProxyEgressPort  = 15001
      ProxyIngressPort = 15000
    }
  }
}

resource "aws_ecs_service" "service" {
  name            = var.service_name
  cluster         = var.ecs_cluster_id
  task_definition = aws_ecs_task_definition.task_definition.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  force_new_deployment = true

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = var.security_groups
    assign_public_ip = true
  }

  service_registries {
    registry_arn   = aws_service_discovery_service.service.arn
    container_name = var.service_name
  }
}
