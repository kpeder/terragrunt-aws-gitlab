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

dependency "custom_vpc" {
  config_path  = find_in_parent_folders(local.env.dependencies.custom_vpc_dependency_path)
  mock_outputs = local.env.dependencies.custom_vpc_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

terraform {
  source = "git::git@github.com:terraform-aws-modules/terraform-aws-security-group?ref=${local.versions.aws_module_sg}"
}

inputs = {
  description         = local.inputs.description
  egress_cidr_blocks  = local.inputs.egress_cidr_blocks
  egress_rules        = local.inputs.egress_rules
  ingress_cidr_blocks = local.inputs.ingress_cidr_blocks
  ingress_rules       = local.inputs.ingress_rules
  name                = format("%s-%s-%s", local.platform.prefix, local.env.environment, local.inputs.name)
  tags                = merge(local.env.labels, local.inputs.labels)
  vpc_id              = dependency.custom_vpc.outputs.vpc_id
}
