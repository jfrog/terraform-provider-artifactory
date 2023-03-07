---
subcategory: "Configuration"
---
# Artifactory LDAP Setting Resource

This resource can be used to manage Artifactory's LDAP settings for user authentication.

When specified LDAP setting is active, Artifactory first attempts to authenticate the user against the LDAP server.
If LDAP authentication fails, it then tries to authenticate via its internal database.

~>The `artifactory_ldap_setting` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
# Configure Artifactory LDAP setting
resource "artifactory_ldap_setting" "ldap_name" {
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
Note: `Key` argument has to match to the resource name.   
Reference Link: [JFrog LDAP](https://www.jfrog.com/confluence/display/JFROG/LDAP)

## Argument Reference

The following arguments are supported:

* `key`                          - (Required) The unique ID of the LDAP setting.
* `enabled`                      - (Optional) When set, these settings are enabled. Default value is `true`.
* `ldap_url`                     - (Required) Location of the LDAP server in the following format: ldap://myserver:myport/dc=sampledomain,dc=com. The URL should include the base DN used to search for and/or authenticate users.
* `user_dn_pattern`              - (Optional) A DN pattern used to log users directly in to the LDAP database. This pattern is used to create a DN string for "direct" user authentication, and is relative to the base DN in the LDAP URL. The pattern argument {0} is replaced with the username at runtime. This only works if anonymous binding is allowed and a direct user DN can be used (which is not the default case for Active Directory). For example: uid={0},ou=People. Default value is blank/empty.
  - Note: LDAP settings should provide a userDnPattern or a searchFilter (or both).
* `email_attribute`              - (Optional) An attribute that can be used to map a user's email address to a user created automatically in Artifactory. Default value is `mail`.
  - Note: If blank/empty string input was set for email_attribute, Default value `mail` takes effect. This is to match with Artifactory behavior.  
* `auto_create_user`             - (Optional) When set, the system will automatically create new users for those who have logged in using LDAP, and assign them to the default groups.  Default value is `true`.
* `ldap_poisoning_protection`    - (Optional) Protects against LDAP poisoning by filtering out users exposed to vulnerabilities.  Default value is `true`.
* `allow_user_to_access_profile` - (Optional) When set, users created after logging in using LDAP will be able to access their profile page.  Default value is `false`.
* `paging_support_enabled`       - (Optional) When set, supports paging results for the LDAP server. This feature requires that the LDAP Server supports a PagedResultsControl configuration.  Default value is `true`.
* `search_filter`                - (Optional) A filter expression used to search for the user DN that is used in LDAP authentication. This is an LDAP search filter (as defined in 'RFC 2254') with optional arguments. In this case, the username is the only argument, denoted by '{0}'. Possible examples are: uid={0}) - this would search for a username match on the uid attribute. Authentication using LDAP is performed from the DN found if successful. Default value is blank/empty.
  - Note: LDAP settings should provide a userDnPattern or a searchFilter (or both)
* `search_base`                  - (Optional) The Context name in which to search relative to the base DN in the LDAP URL. Multiple search bases may be specified separated by a pipe ( | ).
* `search_sub_tree`              - (Optional) When set, enables deep search through the sub-tree of the LDAP URL + Search Base.  Default value is `true`.
* `manager_dn`                   - (Optional) The full DN of a user with permissions that allow querying the LDAP server. When working with LDAP Groups, the user should have permissions for any extra group attributes such as memberOf.
* `manager_password`             - (Optional) The password of the user binding to the LDAP server when using "search" authentication.

## Import

LDAP setting can be imported using the key, e.g.

```
$ terraform import artifactory_ldap_setting.ldap_name ldap_name
```
