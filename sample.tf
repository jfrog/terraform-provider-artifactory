# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.8"
    }
  }
}
resource "artifactory_virtual_maven_repository" "foo" {
  key          = "maven-virt-repo"
  package_type = "maven"
  repo_layout_ref = "maven-2-default"
  repositories = []
  description = "A test virtual repo"
  notes = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
  force_maven_authentication = true
  pom_repository_references_cleanup_policy = "discard_active_reference"
}