resource "aws_lb" "flowsys" {
  name               = "flowsys-lb"
  internal           = false
  load_balancer_type = "application"
  subnets            = data.aws_subnet_ids.public.ids
  security_groups    = [aws_security_group.alb.id]

  tags = {
    Name = "flowsys-lb"
  }
}

resource "aws_lb_target_group" "flowsys" {
  name                 = "flowsys"
  port                 = var.application_port
  protocol             = "HTTP"
  vpc_id               = var.vpc_id
  target_type          = "ip"
  deregistration_delay = 10

  health_check {
    protocol            = "HTTP"
    healthy_threshold   = 3
    unhealthy_threshold = 3
    interval            = 10
    path                = "/stats"
    matcher             = "200"
    port                = 9901
  }
}

resource "aws_lb_listener" "flowsys" {
  load_balancer_arn = aws_lb.flowsys.arn
  port              = "443"
  protocol          = "HTTPS"

  ssl_policy      = "ELBSecurityPolicy-2016-08"
  certificate_arn = aws_acm_certificate_validation.flowsys.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.flowsys.arn
  }
}
