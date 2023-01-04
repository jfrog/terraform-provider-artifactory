# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "6.22.3"
    }
  }
}

provider "artifactory" {
  //  supply ARTIFACTORY_ACCESS_TOKEN / JFROG_ACCESS_TOKEN / ARTIFACTORY_API_KEY and ARTIFACTORY_URL / JFROG_URL as env vars
}

#resource "artifactory_virtual_pypi_repository" "foo-pypi" {
#  key              = "foo-pypi"
#  repositories     = []
#  //description      = "A test virtual repo"
#  notes            = "Internal description"
#  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
#  excludes_pattern = "com/google/**"
#}

resource "artifactory_remote_rpm_repository" "my-remote-rpm" {
  key                     = "my-remote-rpm"
  url                     = "http://mirror.centos.org/centos/"
  //username          = "admin"
  //includes_pattern  = "com/jfrog/**,cloud/jfrog/**"
  //description             = "eee"
  remote_repo_layout_ref  = "build-default"
  //xray_index        = true
}

#resource "artifactory_remote_rpm_repository" "to-import" {
#  key               = "to-import"
#  url               = "http://mirror.centos.org/centos/"
#  //username          = "admin"
#  //includes_pattern  = "com/jfrog/**,cloud/jfrog/**"
#  description       = "qq"
#  //xray_index        = true
#  remote_repo_layout_ref  = "simple-default"
#}

#resource "artifactory_remote_bower_repository" "my-remote-bower" {
#  key              = "my-remote-bower"
#  url              = "https://github.com/"
#  vcs_git_provider = "GITHUB"
#}