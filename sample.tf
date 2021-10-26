# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    xray = {
      source  = "registry.terraform.io/jfrog/xray"
      version = "0.0.1"
    }
  }
}

provider "xray" {
  //  supply ARTIFACTORY_URL and ARTIFACTORY_ACCESS_TOKEN as env vars
}
resource "xray_security_policy" "test" {
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

//resource "xray_watch" "test" {
//  name        = "watch-npm-local-repo"
//  description = "apply a severity-based policy to the npm local repo"
//
//  resources {
//    type       = "repository"
//    name       = "npm-local"
//    bin_mgr_id = "example-com-artifactory-instance"
//    repo_type  = "local"
//    filters {
//      type  = "package-type"
//      value = "Npm"
//    }
//  }
//
//  resources {
//    type       = "repository"
//    name       = "npm-remote"
//    bin_mgr_id = "default"
//    repo_type  = "remote"
//
//    filters {
//      type  = "package-type"
//      value = "Npm"
//    }
//  }
//
//  assigned_policies {
//    name = xray_policy.test.name
//    type = "security"
//  }
//}
