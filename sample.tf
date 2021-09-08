# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source = "registry.terraform.io/jfrog/artifactory"
      version = "2.2.16"
    }
  }
}
variable "supported_repo_types" {
  type = list(string)
  default = [
    "alpine",
    "bower",
    "cargo", // xray refuses to watch these. They also require a mandatory field we can't currently support
    "chef",
    "cocoapods",
    "composer",
    "conan",
    "conda",
    "cran",
    "debian",
    "docker",
    "gems",
    "generic",
    "gitlfs",
    "go",
    "gradle",
    "helm",
    "ivy",
    "maven",
    "npm",
    "nuget",
    "opkg",
    "p2",
    "puppet",
    "pypi",
    // type 'yum' is not to be supported, as this is really of type 'rpm'. When 'yum' is used on create, RT will
    // respond with 'rpm' and thus confuse TF into think there has been a state change.
    "rpm",
    "sbt",
    "vagrant",
    "vcs",
  ]
}
resource "artifactory_local_repository" "local" {
  count = length(var.supported_repo_types)
  key =
  package_type = var.supported_repo_types[count.index]
  xray_index = false
  description = "hello ${var.supported_repo_types[count.index]}-local"
}
provider "artifactory" {
//  supply ARTIFACTORY_USERNAME, _PASSWORD and _URL as env vars
}
resource "artifactory_remote_repository" "npm-remote" {
  key = "npm-remote"
  package_type = "npm"
  url = "https://registry.npmjs.org"
  xray_index = true
}

resource "artifactory_xray_policy" "test" {
  name = "test-policy-name-severity"
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
  name = "watch-npm-local-repo"
  description = "apply a severity-based policy to the npm local repo"

  resources {
    type = "repository"
    name = "npm-local"
    bin_mgr_id = "example-com-artifactory-instance"
    repo_type = "local"
    filters {
      type = "package-type"
      value = "Npm"
    }
  }

  resources {
    type = "repository"
    name = artifactory_remote_repository.npm-remote.key
    bin_mgr_id = "default"
    repo_type = "remote"

    filters {
      type = "package-type"
      value = "Npm"
    }
  }

  assigned_policies {
    name = artifactory_xray_policy.test.name
    type = "security"
  }
}
