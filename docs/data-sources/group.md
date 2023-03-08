# Artifactory Group Data Source

Provides an Artifactory group datasource. This can be used to read the configuration of groups in artifactory.

## Example Usage

```hcl
#
data "artifactory_group" "my_group" {
  name  = "my_group"
  include_users = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the group.
* `include_users` - (Optional) Determines if the group's associated user list will return as an attribute. Default is `false`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `description`         - A description for the group
* `external_id`         - New external group ID used to configure the corresponding group in Azure AD.
* `auto_join`           - When this parameter is set, any new users defined in the system are automatically assigned to this group.
* `admin_privileges`    - Any users added to this group will automatically be assigned with admin privileges in the system.
* `realm`               - The realm for the group.
* `realm_attributes`    - The realm attributes for the group.
* `users_names`         - List of users assigned to the group. Set include_users to `true` to retrieve this list.
* `watch_manager`       - When this override is set, User in the group can manage Xray Watches on any resource type. Default value is `false`.
* `policy_manager`      - When this override is set, User in the group can set Xray security and compliance policies. Default value is `false`.
* `reports_manager`     - When this override is set, User in the group can manage Xray Reports on any resource type. Default value is `false`.
