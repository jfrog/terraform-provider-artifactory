data "artifactory_repositories" "all-alpine-local" {
  repository_type = "local"
  package_type    = "alpine"
}