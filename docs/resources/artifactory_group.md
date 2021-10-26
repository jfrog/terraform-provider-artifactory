# Artifactory Group Resource

Provides an Artifactory group resource. This can be used to create and manage Artifactory groups.

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

* `name`                - (Required) Name of the group
* `description`         - (Optional) A description for the group
* `auto_join`           - (Optional) When this parameter is set, any new users defined in the system are automatically assigned to this group.
* `admin_privileges`    - (Optional) Any users added to this group will automatically be assigned with admin privileges in the system.
* `realm`               - (Optional) The realm for the group.
* `realm_attributes`    - (Optional) The realm attributes for the group.
* `users_names`         - (Optional) List of users assigned to the group. If missing or empty, tf will not manage group membership
* `detach_all_users`    - (Optional) When this override is set, an empty or missing usernames array will detach all users from the group

## Import

Groups can be imported using their name, e.g.

```
$ terraform import artifactory_group.terraform-group mygroup
```
