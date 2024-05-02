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

terraform {
  source = "git::git@github.com:terraform-aws-modules/terraform-aws-acm?ref=${local.versions.aws_module_acm}"
}

inputs = {
  domain_name         = local.env.dns.domain

  subject_alternative_names = [
    format("%s.%s", local.inputs.name, local.env.dns.domain)
  ]

  tags                = merge(local.env.labels, local.inputs.labels)
  validation_method   = "DNS"
  wait_for_validation = local.inputs.wait_for_validation
  zone_id             = local.env.dns.zone_id
}
