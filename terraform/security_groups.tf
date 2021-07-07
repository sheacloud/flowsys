resource "aws_security_group" "alb" {
  name        = "flowsys-alb-sg"
  description = "Allow external HTTPS connections"
  vpc_id      = var.vpc_id

  tags = {
    "Name" = "flowsys-alb-sg"
  }
}

resource "aws_security_group_rule" "alb_https_ingress" {
  security_group_id = aws_security_group.alb.id
  type              = "ingress"
  from_port         = 443
  to_port           = 443
  protocol          = "tcp"
  cidr_blocks       = var.alb_ingress_cidrs
}

resource "aws_security_group_rule" "alb_virtual_gateway_egress" {
  security_group_id        = aws_security_group.alb.id
  type                     = "egress"
  from_port                = var.application_port
  to_port                  = var.application_port
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.virtual_gateway.id
}

resource "aws_security_group_rule" "alb_virtual_gateway_egress_health" {
  security_group_id        = aws_security_group.alb.id
  type                     = "egress"
  from_port                = 9901
  to_port                  = 9901
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.virtual_gateway.id
}


resource "aws_security_group" "virtual_gateway" {
  name        = "flowsys-virtual-gateway-sg"
  description = "Allow connections"
  vpc_id      = var.vpc_id

  tags = {
    Name = "flowsys-virtual-gateway-sg"
  }
}

resource "aws_security_group_rule" "virtual_gateway_ingress" {
  security_group_id        = aws_security_group.virtual_gateway.id
  type                     = "ingress"
  from_port                = var.application_port
  to_port                  = var.application_port
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.alb.id
}

resource "aws_security_group_rule" "virtual_gateway_health" {
  security_group_id        = aws_security_group.virtual_gateway.id
  type                     = "ingress"
  from_port                = 9901
  to_port                  = 9901
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.alb.id
}

resource "aws_security_group_rule" "virtual_gateway_egress" {
  security_group_id = aws_security_group.virtual_gateway.id
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}


resource "aws_security_group" "ingestion" {
  name        = "flowsys-ingestion-sg"
  description = "Allow connections"
  vpc_id      = var.vpc_id

  tags = {
    Name = "flowsys-ingestion-sg"
  }
}

resource "aws_security_group_rule" "ingestion_ingress" {
  security_group_id        = aws_security_group.ingestion.id
  type                     = "ingress"
  from_port                = var.application_port
  to_port                  = var.application_port
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.virtual_gateway.id
}

resource "aws_security_group_rule" "ingestion_egress" {
  security_group_id = aws_security_group.ingestion.id
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}


resource "aws_security_group" "ui" {
  name        = "flowsys-ui-sg"
  description = "Allow connections"
  vpc_id      = var.vpc_id

  tags = {
    Name = "flowsys-ui-sg"
  }
}

resource "aws_security_group_rule" "ui_ingress" {
  security_group_id        = aws_security_group.ui.id
  type                     = "ingress"
  from_port                = var.application_port
  to_port                  = var.application_port
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.virtual_gateway.id
}

resource "aws_security_group_rule" "ui_egress" {
  security_group_id = aws_security_group.ui.id
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}

resource "aws_security_group" "query" {
  name        = "flowsys-query-sg"
  description = "Allow connections"
  vpc_id      = var.vpc_id

  tags = {
    Name = "flowsys-query-sg"
  }
}

resource "aws_security_group_rule" "query_ingress" {
  security_group_id        = aws_security_group.query.id
  type                     = "ingress"
  from_port                = var.application_port
  to_port                  = var.application_port
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.virtual_gateway.id
}

resource "aws_security_group_rule" "query_egress" {
  security_group_id = aws_security_group.query.id
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = "-1"
  cidr_blocks       = ["0.0.0.0/0"]
}
