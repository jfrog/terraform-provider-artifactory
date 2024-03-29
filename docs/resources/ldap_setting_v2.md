---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "artifactory_ldap_setting_v2 Resource - terraform-provider-artifactory"
subcategory: "Configuration"
---
# Artifactory Ldap Setting v2 Resource

Provides an Artifactory LDAP Setting resource. 

This resource can be used to manage Artifactory's LDAP settings for user authentication.

When specified LDAP setting is active, Artifactory first attempts to authenticate the user against the LDAP server.
If LDAP authentication fails, it then tries to authenticate via its internal database.

[API documentation](https://jfrog.com/help/r/jfrog-rest-apis/ldap-setting), [general documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/ldap).

## Example Usage

```terraform
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
```

## Argument reference

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) Ldap setting name.
- `ldap_url` (String) Location of the LDAP server in the following format: `ldap://myldapserver/dc=sampledomain,dc=com`

### Optional

- `allow_user_to_access_profile` (Boolean) Auto created users will have access to their profile page and will be able to perform actions such as generating an API key. Default value is `false`.
- `auto_create_user` (Boolean) When set, users are automatically created when using LDAP. Otherwise, users are transient and associated with auto-join groups defined in Artifactory. Default value is `true`.
- `email_attribute` (String) An attribute that can be used to map a user's email address to a user created automatically in Artifactory. Default value is`mail`.
- `enabled` (Boolean) Flag to enable or disable the ldap setting. Default value is `true`.
- `ldap_poisoning_protection` (Boolean) When this is set to `true`, an empty or missing usernames array will detach all users from the group.
- `manager_dn` (String) The full DN of the user that binds to the LDAP server to perform user searches. Only used with `search` authentication.
- `manager_password` (String, Sensitive) The password of the user that binds to the LDAP server to perform the search. Only used with `search` authentication.
- `paging_support_enabled` (Boolean) When set, supports paging results for the LDAP server. This feature requires that the LDAP server supports a PagedResultsControl configuration. Default value is `true`.
- `search_base` (String) A context name to search in relative to the base DN of the LDAP URL. For example, 'ou=users' With the LDAP Group Add-on enabled, it is possible to enter multiple search base entries separated by a pipe ('|') character.
- `search_filter` (String) A filter expression used to search for the user DN used in LDAP authentication. This is an LDAP search filter (as defined in 'RFC 2254') with optional arguments. In this case, the username is the only argument, and is denoted by '{0}'. Possible examples are: (uid={0}) - This searches for a username match on the attribute. Authentication to LDAP is performed from the DN found if successful.
- `search_sub_tree` (Boolean) When set, enables deep search through the sub tree of the LDAP URL + search base. Default value is `true`.
- `user_dn_pattern` (String) A DN pattern that can be used to log users directly in to LDAP. This pattern is used to create a DN string for 'direct' user authentication where the pattern is relative to the base DN in the LDAP URL. The pattern argument {0} is replaced with the username. This only works if anonymous binding is allowed and a direct user DN can be used, which is not the default case for Active Directory (use User DN search filter instead). Example: uid={0},ou=People. Default value is blank/empty.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

```shell
terraform import artifactory_ldap_setting_v2.ldap ldap1
```
