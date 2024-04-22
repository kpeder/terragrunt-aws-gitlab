

locals {
  inputs = yamldecode(file("./inputs.yaml"))
}

module "instance_connect" {
  source = "../../."

  subnet_id = module.vpc.public_subnets[0]
  security_group_ids = length(local.inputs.security_group_ids) == 0 ? null : local.inputs.security_group_ids
}

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  azs  = local.inputs.azs
  cidr = local.inputs.cidr
  name = local.inputs.name
  public_subnets = local.inputs.public_subnets

}

output "subnet_id" {
  value = module.instance_connect.subnet_id
}

output "vpc_id" {
  value = module.instance_connect.vpc_id
}
