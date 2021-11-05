# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.17"
    }
  }
}
resource "artifactory_group" "test" {
  name             = "test"
  description      = "test"
  admin_privileges = false
  auto_join        = false
}