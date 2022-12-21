---
subcategory: "Security"
---
# Artifactory Group Resource

Provides an Artifactory group resource. This can be used to create and manage Artifactory groups.

## Example Usage

```hcl
# Create a new Artifactory group called terraform
resource "artifactory_group" "test-group" {
  name             = "terraform"
  description      = "test group"
  external_id      = "00628948-b509-4362-aa73-380c4dbd2a44"
  admin_privileges = false
  users_names      = ["foobar"]
}
```

## Argument Reference

The following arguments are supported:

* `name`                - (Required) Name of the group
* `description`         - (Optional) A description for the group
* `external_id`         - (Optional) New external group ID used to configure the corresponding group in Azure AD.
* `auto_join`           - (Optional) When this parameter is set, any new users defined in the system are automatically assigned to this group.
* `admin_privileges`    - (Optional) Any users added to this group will automatically be assigned with admin privileges in the system.
* `realm`               - (Optional) The realm for the group.
* `realm_attributes`    - (Optional) The realm attributes for the group.
* `users_names`         - (Optional) List of users assigned to the group. If not set or empty, Terraform will not manage group membership
* `detach_all_users`    - (Optional) When this is set to `true`, an empty or missing usernames array will detach all users from the group
* `watch_manager`       - (Optional) When this override is set, User in the group can manage Xray Watches on any resource type. Default value is `false`.
* `policy_manager`      - (Optional) When this override is set, User in the group can set Xray security and compliance policies. Default value is `false`.
* `reports_manager`     - (Optional) When this override is set, User in the group can manage Xray Reports on any resource type. Default value is `false`.

## Import

Groups can be imported using their name, e.g.

```
$ terraform import artifactory_group.terraform-group mygroup
```


## Managed vs Unmanaged Group Membership

Terraform does not distinguish between an absent `users_names` attribute and setting to an empty array (i.e. length of 0).

To prevent accidental deletion of existing membership, the default was chosen to mean that Terraform does not manage membership and that to detach all users would require an explicit bool.

~>When moving from managed group membership to unmanaged the Terraform plan will show the users previously in the array
being removed from terraform state, but it will not actually delete any members (unless `detach_all_users` is set to `true`).

Also see [our recommendation](guides/user_group.md) on how to manage user-group relationship.
