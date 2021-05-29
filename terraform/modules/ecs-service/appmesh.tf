resource "aws_appmesh_virtual_node" "service" {
  name      = var.service_name
  mesh_name = var.appmesh_name

  spec {
    dynamic "backend" {
      for_each = var.backends
      content {
        virtual_service {
          virtual_service_name = "${backend.value}.${var.cloud_map_namespace_name}"
        }
      }
    }

    listener {
      port_mapping {
        port     = 8080
        protocol = "http"
      }
    }

    service_discovery {
      aws_cloud_map {
        service_name   = var.service_name
        namespace_name = var.cloud_map_namespace_name
      }
    }

    logging {
      access_log {
        file {
          path = "/dev/stdout"
        }
      }
    }
  }
}


resource "aws_appmesh_virtual_service" "service" {
  name      = "${var.service_name}.${var.cloud_map_namespace_name}"
  mesh_name = var.appmesh_name

  spec {
    provider {
      virtual_node {
        virtual_node_name = aws_appmesh_virtual_node.service.name
      }
    }
  }
}
