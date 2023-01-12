terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "5.0.0"
    }
  }
}

provider "artifactory" {
  //  supply ARTIFACTORY_ACCESS_TOKEN / JFROG_ACCESS_TOKEN / ARTIFACTORY_API_KEY and ARTIFACTORY_URL / JFROG_URL as env vars
}

resource "artifactory_user" "new_user" {
  name   = "new_user"
  email  = "new_user@somewhere.com"
  groups = ["readers"]
}

resource "artifactory_scoped_token" "user" {
  username = artifactory_user.new_user.name
}

resource "artifactory_local_npm_repository" "alexh-npm-local" {
  key         = "alexh-npm-local-key"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_npm_repository" "alexh-npm-local-2" {
  key         = "alexh-npm-local-2-key"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_remote_npm_repository" "alexh-npm-remote" {
  key             = "alexh-npm-remote-key"
  description     = artifactory_user.new_user.name
  url             = "https://registry.npmjs.org/"
  repo_layout_ref = "npm-default"
}

resource "artifactory_remote_npm_repository" "alexh-npm-remote-2" {
  key             = "alexh-npm-remote-2-key"
  url             = "https://registry.npmjs.org/"
  repo_layout_ref = "npm-default"
}

resource "artifactory_virtual_npm_repository" "alexh-npm-virtual" {
  key = "alexh-npm-virtual-key"
  repositories = [
    "${artifactory_local_npm_repository.alexh-npm-local.key}"
  ]
}

resource "artifactory_virtual_npm_repository" "alexh-npm-virtual-2" {
  key         = "alexh-npm-virtual-2-key"
  description = "Foo ${artifactory_local_npm_repository.alexh-npm-local.key} Bar"
  repositories = [
    "${artifactory_local_npm_repository.alexh-npm-local.key}",
    "${artifactory_local_npm_repository.alexh-npm-local-2.key}"
  ]
}
