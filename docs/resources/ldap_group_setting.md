---
subcategory: "Configuration"
---
# Artifactory LDAP Group Setting Resource

This resource can be used to manage Artifactory's LDAP Group settings for user authentication.

LDAP Groups Add-on allows you to synchronize your LDAP groups with the system and leverage your existing organizational
structure for managing group-based permissions.

~>The `artifactory_ldap_group_setting` resource utilizes endpoints which are blocked/removed in SaaS environments (i.e. in Artifactory online), rendering this resource incompatible with Artifactory SaaS environments.

## Example Usage

```hcl
# Configure Artifactory LDAP setting
resource "artifactory_ldap_group_setting" "ldap_group_name" {
  name                    = "ldap_group_name"
  ldap_setting_key        = "ldap_name"
  group_base_dn           = ""
  group_name_attribute    = "cn"
  group_member_attribute  = "uniqueMember"
  sub_tree                = true
  filter                  = "(objectClass=groupOfNames)"
  description_attribute   = "description"
  strategy                = "STATIC"
}
```
Note: `Name` argument has to match to the resource name.   
Reference Link: [JFrog LDAP](https://www.jfrog.com/confluence/display/JFROG/LDAP)

## Argument Reference

The following arguments are supported:

* `name`                          - (Required) Ldap group setting name.
* `ldap_setting_key`              - (Required) The LDAP setting key you want to use for group retrieval. The value for this field corresponds to 'enabledLdap' field of the ldap group setting XML block of system configuration.
* `group_base_dn`                 - (Optional) A search base for group entry DNs, relative to the DN on the LDAP server’s URL (and not relative to the LDAP Setting’s “Search Base”). Used when importing groups.
* `group_name_attribute`          - (Required) Attribute on the group entry denoting the group name. Used when importing groups.
* `group_member_attribute`        - (Required) A multi-value attribute on the group entry containing user DNs or IDs of the group members (e.g., uniqueMember,member).
* `sub_tree`                      - (Optional) When set, enables deep search through the sub-tree of the LDAP URL + Search Base. True by default.
* `filter`                        - (Required) The LDAP filter used to search for group entries. Used for importing groups.
* `description_attribute`         - (Required) An attribute on the group entry which denoting the group description. Used when importing groups.
* `strategy`                      - (Required) The JFrog Platform Deployment (JPD) supports three ways of mapping groups to LDAP schemas:
  - STATIC: Group objects are aware of their members, however, the users are not aware of the groups they belong to. Each group object such as groupOfNames or groupOfUniqueNames holds its respective member attributes, typically member or uniqueMember, which is a user DN.
  - DYNAMIC: User objects are aware of what groups they belong to, but the group objects are not aware of their members. Each user object contains a custom attribute, such as group, that holds the group DNs or group names of which the user is a member.
  - HIERARCHICAL: The user's DN is indicative of the groups the user belongs to by using group names as part of user DN hierarchy. Each user DN contains a list of ou's or custom attributes that make up the group association. For example, uid=user1,ou=developers,ou=uk,dc=jfrog,dc=org indicates that user1 belongs to two groups: uk and developers.

## Import

LDAP Group setting can be imported using the key, e.g.

```
$ terraform import artifactory_ldap_group_setting.ldap_group_name ldap_group_name
```
