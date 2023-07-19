resource "artifactory_ldap_setting_v2" "ldap_name" {
  key                           = "ldap_name"
  enabled                       = true
  ldap_url                      = "ldap://ldap_server_url"
  user_dn_pattern               = "uid={0},ou=People"
  email_attribute               = "mail"
  auto_create_user              = true
  ldap_poisoning_protection     = true
  allow_user_to_access_profile  = false
  paging_support_enabled        = false
  search_filter                 = "(uid={0})"
  search_base                   = "ou=users"
  search_sub_tree               = true
  manager_dn                    = "mgr_dn"
  manager_password              = "mgr_passwd_random"
}