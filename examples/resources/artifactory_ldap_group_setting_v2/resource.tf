resource "artifactory_ldap_group_setting_v2" "ldap_group_name" {
  name                   = "ldap_group_name"
  enabled_ldap           = "ldap_name"
  group_base_dn          = "CN=Users,DC=MyDomain,DC=com"
  group_name_attribute   = "cn"
  group_member_attribute = "uniqueMember"
  sub_tree               = true
  force_attribute_search = false
  filter                 = "(objectClass=groupOfNames)"
  description_attribute  = "description"
  strategy               = "STATIC"
}