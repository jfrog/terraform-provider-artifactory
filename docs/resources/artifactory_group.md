# Artifactory Group Resource

Provides an Artifactory user-group resource. This can be used to create and manage Artifactory user-group.

## Example Usage

```hcl
# Create a new Artifactory group called terraform
resource "artifactory_group" "test-group" {
  name             = "terraform"
  description      = "test group"
  admin_privileges = false
  users_names      = [ "foobar" ]
}
```

## Argument Reference

The following arguments are supported:

* `name`                - (Required) Name of the user-group.

* `description`         - (Optional) A description for the user-group.

* `auto_join`           - (Optional) A value of `true` means that new users defined in the system are automatically assigned to this user-group. A value of `false` means that new users defined in the system are NOT assigned to this user-group. The default value is `false`.

* `admin_privileges`    - (Optional) A value of `true` means that any users added to this group will automatically be assigned with admin privileges in the system. A value of `false` means that users added to this group will NOT be automatically assigned with admin privileges. The default value is `false`.

* `realm`               - (Optional) The realm for the group.

* `realm_attributes`    - (Optional) The realm attributes for the group.

* `users_names`         - (Optional) List of users assigned to the group. If missing or empty, Terraform will not manage group membership.

* `detach_all_users`    - (Optional) A value of `true` means that an empty or missing `users_names` array will detach all users from the group. A value of `false` means that an empty or missing `users_names` array will NOT detach all users from the group. The default value is `false`.

## Import

Groups can be imported using their name.

```
terraform import artifactory_group.terraform-group mygroup
```

## Managed vs Unmanaged Group Membership

Terraform does not distinguish between an _absent_ vs _empty_ `users_names` array. To prevent accidental deletion of existing membership, the default value means that Terraform does not manage membership. To detach all users would require an explicit boolean `true`.

~> **Note:** When moving from managed group membership to unmanaged, the `terraform plan` will show the users previously in the array being removed from `terraform.tfstate` file, but it will not actually delete any members from the Artifactory system.
