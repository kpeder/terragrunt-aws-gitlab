# Terragrunt will copy the Terraform configurations specified by the source parameter, along with any files in the
# working directory, into a temporary folder, and execute your Terraform commands in that folder.

# Include all settings from the root terragrunt.hcl file
include {
  path = find_in_parent_folders("aws_gitlab_terragrunt.hcl")
}

# Resources should not be destroyed without careful consideration of effects
prevent_destroy = false

locals {
  env      = yamldecode(file(find_in_parent_folders("env.yaml")))
  inputs   = yamldecode(file("inputs.yaml"))
  platform = fileexists(find_in_parent_folders("local.aws.yaml")) ? yamldecode(file(find_in_parent_folders("local.aws.yaml"))) : yamldecode(file(find_in_parent_folders("aws.yaml")))
  region   = yamldecode(file(find_in_parent_folders("region.yaml")))
  versions = yamldecode(file(find_in_parent_folders("versions.yaml")))
}

terraform {
  source = "git::git@github.com:terraform-aws-modules/terraform-aws-vpc?ref=${local.versions.aws_module_vpc}"
}

inputs = {
  azs                           = [for zone in local.inputs.zones: format("%s%s", local.region.location, zone)]
  cidr                          = local.inputs.cidr
  create_igw                    = local.inputs.internet.deploy_gateway
  enable_dns_hostnames          = local.inputs.dns.hostnames
  enable_dns_support            = local.inputs.dns.support
  name                          = format("%s-%s-%s", local.platform.prefix, local.env.environment, local.inputs.name)

  intra_subnets                 = local.inputs.subnets.intra
  private_subnets               = local.inputs.subnets.private
  public_subnets                = local.inputs.subnets.public

  default_dedicated_network_acl = local.inputs.network_acls.default.managed
  default_inbound_acl_rules     = local.inputs.network_acls.default.inbound_rules
  default_outbound_acl_rules    = local.inputs.network_acls.default.outbound_rules

  intra_dedicated_network_acl   = local.inputs.network_acls.intra.managed
  intra_inbound_acl_rules       = local.inputs.network_acls.intra.inbound_rules
  intra_outbound_acl_rules      = local.inputs.network_acls.intra.outbound_rules

  public_dedicated_network_acl  = local.inputs.network_acls.public.managed
  public_inbound_acl_rules      = local.inputs.network_acls.public.inbound_rules
  public_outbound_acl_rules     = local.inputs.network_acls.public.outbound_rules

  private_dedicated_network_acl = local.inputs.network_acls.private.managed
  private_inbound_acl_rules     = local.inputs.network_acls.private.inbound_rules
  private_outbound_acl_rules    = local.inputs.network_acls.private.outbound_rules

  enable_nat_gateway            = local.inputs.nat.deploy_gateways
  one_nat_gateway_per_az        = local.inputs.nat.gateway_per_zone
  single_nat_gateway            = local.inputs.nat.global_gateway

  enable_vpn_gateway            = local.inputs.vpn.deploy_gateway
  vpn_gateway_az                = format("%s%s", local.region.location, local.region.zone_preference)

  tags                          = merge(local.env.labels, local.inputs.labels)
}
