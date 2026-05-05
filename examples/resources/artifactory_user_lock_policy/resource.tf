resource "artifactory_user_lock_policy" "my-user-lock-policy" {
  name           = "my-user-lock-policy"
  enabled        = true
  login_attempts = 10
}