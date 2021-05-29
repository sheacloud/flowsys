variable "ecs_cluster_id" {
  type = string
}

variable "service_name" {
  type = string
}

variable "cloud_map_namespace_id" {
  type = string
}

variable "cloud_map_namespace_name" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "security_groups" {
  type = list(string)
}

variable "appmesh_name" {
  type = string
}

variable "backends" {
  type    = list(any)
  default = []
}

variable "application_port" {
  type = number
}

variable "envoy_image" {
  type = string
}
