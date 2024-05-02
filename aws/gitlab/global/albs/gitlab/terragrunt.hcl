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
  versions = yamldecode(file(find_in_parent_folders("versions.yaml")))
}

dependency "custom_vpc" {
  config_path  = find_in_parent_folders(local.env.dependencies.custom_vpc_dependency_path)
  mock_outputs = local.env.dependencies.custom_vpc_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

dependency "gitlab_certificate" {
  config_path  = find_in_parent_folders(local.env.dependencies.gitlab_certificate_dependency_path)
  mock_outputs = local.env.dependencies.gitlab_certificate_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

dependency "gitlab_instance" {
  config_path  = find_in_parent_folders(local.env.dependencies.gitlab_instance_dependency_path)
  mock_outputs = local.env.dependencies.gitlab_instance_mock_outputs

  mock_outputs_allowed_terraform_commands = ["init", "plan", "validate"]
}

terraform {
  source = "git::git@github.com:terraform-aws-modules/terraform-aws-alb?ref=${local.versions.aws_module_alb}"
}

inputs = {
  enable_deletion_protection = local.inputs.deletion_protection

  name    = local.inputs.name
  subnets = dependency.custom_vpc.outputs.public_subnets
  vpc_id  = dependency.custom_vpc.outputs.vpc_id

  listeners                    = merge({
      for k, v in local.inputs.listeners:
        k => merge(v,
          {
            certificate_arn = dependency.gitlab_certificate.outputs.acm_certificate_arn
          }) if contains(keys(v), "certificate_arn")
    },{
      for k, v in local.inputs.listeners:
        k => v if !contains(keys(v), "certificate_arn")
    })
  route53_records              = {
      for v in local.inputs.dns: v.name => merge(v,
        {
          zone_id = local.env.dns.zone_id
        })
    }
  security_group_egress_rules  = { for k, v in local.inputs.egress: k => v }
  security_group_ingress_rules = { for k, v in local.inputs.ingress: k => v }
  target_groups                = {
    for k, v in local.inputs.targets: k => merge(v, 
    {
      target_id = dependency.gitlab_instance.outputs.id
    })
  }

  tags = merge(local.env.labels, local.inputs.labels)
}
