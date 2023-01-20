---
subcategory: "User"
---
# Artifactory Managed User Resource

Provides an Artifactory managed user resource. This can be used to create and maintain Artifactory users. For example, service account where password is known and managed externally.

Unlike `artifactory_unmanaged_user` and `artifactory_user`, the `password` attribute is required and cannot be empty.
Consider using a separate provider to generate and manage passwords. 

~> The password is stored in the Terraform state file. Make sure you secure it, please refer to the official [Terraform documentation](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

## Example Usage

```hcl
# Create a new Artifactory user called terraform
resource "artifactory_managed_user" "test-user" {
  name     = "terraform"
  email    = "test-user@artifactory-terraform.com"
  groups   = ["logged-in-users", "readers"]
  password = "my super secret password"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Username for user.
* `email` - (Required) Email for user.
* `password` - (Required) Password for the user.
* `admin` - (Optional) When enabled, this user is an administrator with all the ensuing privileges. Default value is `false`.
* `profile_updatable` - (Optional) When set, this user can update his profile details (except for the password. Only an administrator can update the password). Default value is `true`.
* `disable_ui_access` - (Optional) When set, this user can only access Artifactory through the REST API. This option cannot be set if the user has Admin privileges. Default value is `true`.
* `internal_password_disabled` - (Optional) When set, disables the fallback of using an internal password when external authentication (such as LDAP) is enabled.
* `groups` - (Optional) List of groups this user is a part of. **Notes:** If this attribute is not specified then user's group membership is set to empty. User will not be part of default "readers" group automatically.

## Import

Users can be imported using their name, e.g.

```
$ terraform import artifactory_managed_user.test-user myusername
```
