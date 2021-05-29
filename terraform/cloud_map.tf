resource "aws_service_discovery_private_dns_namespace" "flowsys" {
  name        = "flowsys.local"
  description = "flowsys"
  vpc         = var.vpc_id
}

resource "aws_service_discovery_service" "gateway" {
  name = "gateway"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.flowsys.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}
