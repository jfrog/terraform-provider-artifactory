provider "xray" {
  url          = "${var.artifactory_url}/xray"
  access_token = var.xray_access_token
}
