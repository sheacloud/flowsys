resource "aws_service_discovery_service" "service" {
  name = var.service_name

  dns_config {
    namespace_id = var.cloud_map_namespace_id

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
