provider "artifactory" {
  url          = "${var.artifactory_url}/artifactory"
  access_token = var.artifactory_access_token
  check_license   = true
}
