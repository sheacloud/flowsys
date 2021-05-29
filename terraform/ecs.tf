resource "aws_ecs_cluster" "flowsys" {
  name = "flowsys"

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 0
    capacity_provider = "FARGATE"
    weight            = 1
  }

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

data "template_file" "gateway_task_definition" {
  template = file("${path.module}/task-definitions/envoy_gateway.json")
  vars = {
    APPMESH_RESOURCE_ARN = "mesh/${aws_appmesh_mesh.flowsys.name}/virtualGateway/${aws_appmesh_virtual_gateway.gateway.name}"
    APPLICATION_PORT     = var.application_port
    ENVOY_IMAGE          = var.envoy_image
  }
}


resource "aws_ecs_task_definition" "envoy_virtual_gateway" {
  family                   = "flowsys-gateway"
  container_definitions    = data.template_file.gateway_task_definition.rendered
  requires_compatibilities = ["FARGATE"]

  network_mode = "awsvpc"

  cpu                = 256
  memory             = 512
  task_role_arn      = aws_iam_role.gateway_ecs_task_role.arn
  execution_role_arn = aws_iam_role.gateway_ecs_task_execution_role.arn
}

resource "aws_ecs_service" "service" {
  name            = "gateway"
  cluster         = aws_ecs_cluster.flowsys.id
  task_definition = aws_ecs_task_definition.envoy_virtual_gateway.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  force_new_deployment = true

  network_configuration {
    subnets          = data.aws_subnet_ids.public.ids
    security_groups  = [aws_security_group.virtual_gateway.id]
    assign_public_ip = true
  }

  service_registries {
    registry_arn   = aws_service_discovery_service.gateway.arn
    container_name = "gateway"
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.flowsys.arn
    container_name   = "gateway"
    container_port   = var.application_port
  }
}
