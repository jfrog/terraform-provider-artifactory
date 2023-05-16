resource "artifactory_group" "test-group" {
  name             = "terraform"
  description      = "test group"
  external_id      = "00628948-b509-4362-aa73-380c4dbd2a44"
  admin_privileges = false
  users_names      = ["foobar"]
}