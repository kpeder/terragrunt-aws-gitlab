# configure gcs bucket dynamically.
remote_state {
  backend = "gcs"
  config = {
    bucket                 = format("%s-%s-terraform-state", local.platform.prefix, format("%s-environment", local.environment.environment))
    prefix                 = path_relative_to_include()
    location               = local.multiregion.region
    project                = local.platform.build_project
    skip_bucket_creation   = false
    skip_bucket_versioning = false
  }
}

locals {
  default_yaml_path = find_in_parent_folders("empty.yaml") # terragrunt function for input search (not implemented).
  platform          = fileexists(find_in_parent_folders("local.gcp.yaml")) ? yamldecode(file(find_in_parent_folders("local.gcp.yaml"))) : yamldecode(file(find_in_parent_folders("gcp.yaml")))
  environment       = yamldecode(file(find_in_parent_folders("env.yaml")))
  multiregion       = yamldecode(file(find_in_parent_folders("reg-multi/region.yaml")))
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
    google = {
      version = "${format("~> %s.0", local.versions.google_provider_version)}"
    }
    google-beta = {
      version = "${format("~> %s.0", local.versions.google_provider_version)}"
    }
  }
}
EOF
}
