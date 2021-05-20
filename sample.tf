# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.2.7"
    }
  }
}
provider "artifactory" {
  url = "http://localhost:8082/"
  username = "admin"
  password = "password"
}
resource "random_id" "randid" {
  count       = 4
  byte_length = 2
}
resource "random_password" "randpass" {
  count       = 10
  length      = 16
  min_lower   = 5
  min_upper   = 5
  min_numeric = 1
  min_special = 1
}
resource "artifactory_user" "user" {
  count    = length(random_password.randpass)
  name     = "terraform${count.index}"
  email    = "test-user@artifactory-terraform.com"
  groups   = ["readers"]
  password = random_password.randpass[count.index].result
}
resource "artifactory_local_repository" "npm-local" {
  key             = "npm-local"
  package_type    = "npm"
  repo_layout_ref = "npm-default"
  xray_index      = true
}

resource "artifactory_remote_repository" "npm-remote" {
  key          = "npm-remote"
  package_type = "npm"
  url          = "https://registry.npmjs.org"
  xray_index   = true
}

resource "artifactory_remote_repository" "icts-p-icts-alpine-generic-remote" {
  key                    = "icts-p-icts-alpine-generic-remote"
  description            = "alpine Repos"
  package_type           = "generic"
  repo_layout_ref        = "simple-default"
  url                    = "http://dl-cdn.alpinelinux.org/alpine"
  propagate_query_params = true
}

resource "artifactory_xray_policy" "test" {
  name        = "test-policy-name-severity"
  description = "test policy description"
  type        = "security"
  rules {
    name     = "rule-name"
    priority = 1
    criteria {
      min_severity = "High"
    }
    actions {
      block_download {
        unscanned = true
        active    = true
      }
    }
  }
}

resource "artifactory_xray_watch" "test" {
  name        = "watch-npm-local-repo"
  description = "apply a severity-based policy to the npm local repo"

  resources {
    type       = "repository"
    name       = artifactory_local_repository.npm-local.key
    bin_mgr_id = "example-com-artifactory-instance"
    repo_type  = "local"
    filters {
      type  = "package-type"
      value = "Npm"
    }
  }

  resources {
    type       = "repository"
    name       = artifactory_remote_repository.npm-remote.key
    bin_mgr_id = "default"
    repo_type  = "remote"

    filters {
      type  = "package-type"
      value = "Npm"
    }
  }

  assigned_policies {
    name = artifactory_xray_policy.test.name
    type = "security"
  }
}
