
module "ingestion" {
  source = "./modules/ecs-service/"

  ecs_cluster_id           = aws_ecs_cluster.flowsys.id
  service_name             = "ingestion"
  application_port         = var.application_port
  cloud_map_namespace_id   = aws_service_discovery_private_dns_namespace.flowsys.id
  cloud_map_namespace_name = aws_service_discovery_private_dns_namespace.flowsys.name
  subnet_ids               = data.aws_subnet_ids.public.ids
  security_groups          = [aws_security_group.ingestion.id]
  appmesh_name             = aws_appmesh_mesh.flowsys.name
  envoy_image              = var.envoy_image
  backends                 = []
}


module "ui" {
  source = "./modules/ecs-service/"

  ecs_cluster_id           = aws_ecs_cluster.flowsys.id
  service_name             = "ui"
  application_port         = var.application_port
  cloud_map_namespace_id   = aws_service_discovery_private_dns_namespace.flowsys.id
  cloud_map_namespace_name = aws_service_discovery_private_dns_namespace.flowsys.name
  subnet_ids               = data.aws_subnet_ids.public.ids
  security_groups          = [aws_security_group.ui.id]
  appmesh_name             = aws_appmesh_mesh.flowsys.name
  envoy_image              = var.envoy_image
  backends                 = []
}

module "query" {
  source = "./modules/ecs-service/"

  ecs_cluster_id           = aws_ecs_cluster.flowsys.id
  service_name             = "query"
  application_port         = var.application_port
  cloud_map_namespace_id   = aws_service_discovery_private_dns_namespace.flowsys.id
  cloud_map_namespace_name = aws_service_discovery_private_dns_namespace.flowsys.name
  subnet_ids               = data.aws_subnet_ids.public.ids
  security_groups          = [aws_security_group.query.id]
  appmesh_name             = aws_appmesh_mesh.flowsys.name
  envoy_image              = var.envoy_image
  backends                 = []
}
