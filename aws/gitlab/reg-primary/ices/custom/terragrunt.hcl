# Terragrunt will copy the Terraform configurations specified by the source parameter, along with any files in the
# working directory, into a temporary folder, and execute your Terraform commands in that folder.

# Include all settings from the root terragrunt.hcl file
include {
  path = find_in_parent_folders("aws_gitlab_terragrunt.hcl")
}

# Resources should not be destroyed without careful consideration of effects
prevent_destroy = false

locals {
  env    = yamldecode(file(find_in_parent_folders("env.yaml")))
  inputs = yamldecode(file("inputs.yaml"))
}

dependency "custom_vpc" {
  config_path  = find_in_parent_folders(local.env.dependencies.custom_vpc_dependency_path)
  mock_outputs = local.env.dependencies.custom_vpc_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

dependency "custom_ice_sg" {
  config_path  = find_in_parent_folders(local.env.dependencies.custom_ice_sg_dependency_path)
  mock_outputs = local.env.dependencies.custom_ice_sg_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

terraform {
  source = "../../../../../modules/terraform-aws-instance-connect"
}

inputs = {
  preserve_client_ip = local.inputs.preserve_client_ip
  security_group_ids = concat(local.inputs.security_group_ids, tolist([dependency.custom_ice_sg.outputs.security_group_id]))
  subnet_id          = dependency.custom_vpc.outputs.public_subnets[0]
  tags               = merge(local.env.labels, local.inputs.labels)
}
