resource "artifactory_managed_user" "test-user" {
  name     = "terraform"
  password = "my super secret password"
  email    = "test-user@artifactory-terraform.com"
  groups   = [ "readers", "logged-in-users"]
}