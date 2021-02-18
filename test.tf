# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.2.6-alpha"
    }
  }
}
provider "artifactory" {
  url = "http://localhost:8082/artifactory"
  username = "admin"
  password = "password"
}

resource "artifactory_user" "test-user" {
  name     = "terraform"
  email    = "test-user@artifactory-terraform.com"
  groups   = ["readers"]
  password = "password1"
}
resource "artifactory_user" "test-user2" {
  name     = "terraform1"
  email    = "test-user@artifactory-terraform.com"
  groups   = ["readers"]
  password = "password1"
}

