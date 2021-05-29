data "aws_subnet_ids" "public" {
  vpc_id = var.vpc_id

  tags = {
    Layer = "public"
  }
}
