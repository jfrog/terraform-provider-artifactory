# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "7.4.4"
    }
  }
}

provider "artifactory" {
  //  supply ARTIFACTORY_ACCESS_TOKEN / JFROG_ACCESS_TOKEN / ARTIFACTORY_API_KEY and ARTIFACTORY_URL / JFROG_URL as env vars
}

resource "artifactory_federated_npm_repository" "terraform-federated-test-npm-repo" {
  key       = "terraform-federated-test-npm-repo"

  member {
    url     = "http://artifactory-2:8081/artifactory/federated-generic-5"
    enabled = true
  }

  member {
    url     = "http://artifactory-2:8081/artifactory/federated-generic-6"
    enabled = true
  }
}