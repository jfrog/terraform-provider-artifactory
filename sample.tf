# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source = "registry.terraform.io/jfrog/artifactory"
      version = "2.3.5"
    }
  }
}
data "artifactory_file" "ac_api_changelog_indexer_code" {
  repository      = "integration-helm"
  path            = "artifactory-11.7.2.tgz"
  output_path     = "${path.cwd}/resources/integration-helm/artifactory-11.7.2.tgz"
  force_overwrite = true
}