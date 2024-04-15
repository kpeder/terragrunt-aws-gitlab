# configure s3 bucket dynamically.
remote_state {
  backend = "s3"
  config = {
    bucket         = format("%s-%s-terraform-state", local.platform.prefix, format("%s-environment", local.environment.environment))
    key            = path_relative_to_include()
    region         = local.region.region
    encrypt        = true
  }
}

locals {
  default_yaml_path = find_in_parent_folders("empty.yaml") # terragrunt function for input search (not implemented).
  platform          = fileexists(find_in_parent_folders("local.aws.yaml")) ? yamldecode(file(find_in_parent_folders("local.aws.yaml"))) : yamldecode(file(find_in_parent_folders("aws.yaml")))
  environment       = yamldecode(file(find_in_parent_folders("env.yaml")))
  region            = yamldecode(file(find_in_parent_folders("reg-primary/region.yaml")))
  versions          = yamldecode(file(find_in_parent_folders("versions.yaml")))
}

terragrunt_version_constraint = "${format("~> %s.0", local.versions.terragrunt_binary_version)}"
terraform_version_constraint  = "${format("~> %s.0", local.versions.terraform_binary_version)}"

generate "provider" {
  path      = "provider_override.tf"
  if_exists = "overwrite"
  contents = <<EOF
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "${format("~> %s.0", local.versions.aws_provider_version)}"
    }
  }
}
EOF
}
