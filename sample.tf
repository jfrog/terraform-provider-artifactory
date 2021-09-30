# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.8"
    }
  }
}
variable "supported_repo_types" {
  type = list(string)
  default = [
    "alpine",
    "bower",
    // xray refuses to cargo. They also require a mandatory field we can't currently support
    //    "cargo",
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
resource "random_id" "randid" {
  byte_length = 16
}


resource "artifactory_local_repository" "local" {
  count = length(var.supported_repo_types)
  key = "${var.supported_repo_types[count.index]}-local"
  package_type = var.supported_repo_types[count.index]
  xray_index = false
  description = "hello ${var.supported_repo_types[count.index]}-local"
}

resource "artifactory_local_repository" "local-rand" {
  count = 100
  key = "foo-${count.index}-local"
  package_type = var.supported_repo_types[random_id.randid.dec % length(var.supported_repo_types)]
  xray_index = true
  description = "hello ${count.index}-local"
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
resource "artifactory_virtual_go_repository" "baz-go" {
  key          = "baz-go"
  package_type = "go"
  repo_layout_ref = "go-default"
  repositories = []
  description = "A test virtual repo"
  notes = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
  external_dependencies_enabled = true
  external_dependencies_patterns = [
    "**/github.com/**",
    "**/go.googlesource.com/**"
  ]
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


