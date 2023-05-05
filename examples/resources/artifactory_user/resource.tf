resource "artifactory_user" "test-user" {
  name                          = "terraform"
  password                      = "my super secret password"
  email                         = "test-user@artifactory-terraform.com"
  admin                         = false
  profile_updatable   		    = true
  disable_ui_access			    = false
  internal_password_disabled 	= false
  groups                        = ["readers", "logged-in-users"]
}