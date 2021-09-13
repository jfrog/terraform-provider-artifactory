# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source = "registry.terraform.io/jfrog/artifactory"
      version = "2.3.5"
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
resource "artifactory_local_repository" "local" {
  count = length(var.supported_repo_types)
  key = "${var.supported_repo_types[count.index]}-local"
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

