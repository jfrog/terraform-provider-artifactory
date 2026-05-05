resource "artifactory_release_bundle_v2_promotion" "my-release-bundle-v2-promotion" {
  name                     = "my-release-bundle-v2-artifacts"
  version                  = "1.0.0"
  keypair_name             = "my-keypair-name"
  environment              = "DEV"
  included_repository_keys = ["commons-qa-maven-local"]
}