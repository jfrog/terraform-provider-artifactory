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
  url = "http://localhost:8082"
  username = "admin"
  password = "password"
}
resource "random_id" randid {
  count = 4
  byte_length = 2
}
resource  random_password randpass {
  count = 10
  length = 16
  min_lower = 5
  min_upper = 5
  min_numeric = 1
  min_special = 1
}
resource "artifactory_user" "user" {
  count = length(random_password.randpass)
  name     = "terraform${count.index}"
  email    = "test-user@artifactory-terraform.com"
  groups   = ["readers"]
  password = random_password.randpass[count.index].result
}

resource "artifactory_remote_repository" "conan-remote" {
  key = "conan-remote"
  package_type = "conan"
  url = "https://conan.bintray.com"
  repo_layout_ref = "conan-default"
  notes = "managed by terraform"
}

resource "artifactory_xray_policy" "test" {
  name  = "test-policy-name"
  description = "test policy description"
  type = "security"

  rules {
    name = "rule-name"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      block_download {
        unscanned = true
        active = true
      }
    }
  }
}

resource "artifactory_xray_watch" "test" {
  name = "test-watch-name"
  description = "test watch description"
  resources {
    type = "all-repos"
    name = "All Repositories"
  }
  assigned_policies {
    name = artifactory_xray_policy.test.name
    type = "security"
  }
  watch_recipients = ["test@example.com"]
}
