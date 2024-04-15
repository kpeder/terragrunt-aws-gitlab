resource "aws_ec2_instance_connect_endpoint" "this" {
  subnet_id = var.subnet_id
  tags      = var.tags
}
