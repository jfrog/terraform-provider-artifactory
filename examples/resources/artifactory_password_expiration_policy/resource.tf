resource "artifactory_password_expiration_policy" "my-password-expiration-policy" {
  name = "my-password-expiration-policy"
  enabled = true
  password_max_age = 120
  notify_by_email = true
}