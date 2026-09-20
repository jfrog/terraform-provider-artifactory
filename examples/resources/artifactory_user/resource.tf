resource "artifactory_user" "test-user" {
  name     = "terraform"
  password = "my super secret password"
  password_policy = {
    uppercase    = 1
    lowercase    = 1
    special_char = 1
    digit        = 1
    length       = 10
  }
  email                      = "test-user@artifactory-terraform.com"
  admin                      = false
  profile_updatable          = true
  disable_ui_access          = false
  internal_password_disabled = false
  groups                     = ["readers", "logged-in-users"]
}