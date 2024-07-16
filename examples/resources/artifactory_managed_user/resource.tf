resource "artifactory_managed_user" "test-user" {
  name     = "terraform"
  password = "my super secret password"
  password_policy = {
    uppercase = 1
    lowercase = 1
    special_char = 1
    digit = 1
    length = 10
  }
  email    = "test-user@artifactory-terraform.com"
  groups   = [ "readers", "logged-in-users"]
}