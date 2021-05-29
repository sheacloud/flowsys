resource "aws_appmesh_mesh" "flowsys" {
  name = "flowsys"
}


resource "aws_appmesh_virtual_gateway" "gateway" {
  name      = "flowsys-virtual-gateway"
  mesh_name = aws_appmesh_mesh.flowsys.name

  spec {
    listener {
      port_mapping {
        port     = var.application_port
        protocol = "http"
      }
    }
  }
}

resource "aws_appmesh_gateway_route" "ingestion" {
  name                 = "ingestion-gateway-route"
  mesh_name            = aws_appmesh_mesh.flowsys.name
  virtual_gateway_name = aws_appmesh_virtual_gateway.gateway.name

  spec {
    http_route {
      action {
        target {
          virtual_service {
            virtual_service_name = module.ingestion.virtual_service_name
          }
        }
      }

      match {
        prefix = "/ingestion"
      }
    }
  }
}
