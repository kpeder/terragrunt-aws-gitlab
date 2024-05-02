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

dependency "gitlab_keypair" {
  config_path  = find_in_parent_folders(local.env.dependencies.gitlab_keypair_dependency_path)
  mock_outputs = local.env.dependencies.gitlab_keypair_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

dependency "gitlab_sg" {
  config_path  = find_in_parent_folders(local.env.dependencies.gitlab_sg_dependency_path)
  mock_outputs = local.env.dependencies.gitlab_sg_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

terraform {
  source = "git::git@github.com:terraform-aws-modules/terraform-aws-ec2-instance?ref=${local.versions.aws_module_ec2}"
}

inputs = {
  ami                         = local.inputs.ami
  associate_public_ip_address = local.inputs.public_ip
  name                        = format("%s-%s-%s", local.platform.prefix, local.env.environment, local.inputs.name)
  zone                        = format("%s%s", local.region.location, local.region.zone_preference)
  instance_type               = local.inputs.type
  key_name                    = dependency.gitlab_keypair.outputs.key_pair_name
  monitoring                  = local.inputs.monitoring
  subnet_id                   = dependency.custom_vpc.outputs.private_subnets[0]
  tags                        = merge(local.env.labels, local.inputs.labels)
  user_data                   = local.inputs.user_data
  vpc_security_group_ids      = tolist([dependency.gitlab_sg.outputs.security_group_id])
}
