# artifactory_user

Provides an Artifactory user resource. This can be used to create and manage Artifactory users.

## Example Usage

```hcl
# Create a new Artifactory user called terraform
resource "artifactory_user" "test-user" {
  name   = "terraform"
  email  = "test-user@artifactory-terraform.com"
  groups = ["logged-in-users", "readers"]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Username for user
* `email` - (Required) Email for user
* `admin` - (Optional) 
* `profile_updatable` - (Optional) When set, this user can update his profile details (except for the password. Only an administrator can update the password).
* `disable_ui_access` - (Optional) When set, this user can only access Artifactory through the REST API. 
* `internal_password_disabled` - (Optional) When set, disables the fallback of using an internal password when external authentication (such as LDAP) is enabled.
* `groups` - (Optional) List of groups this user is a part of

## Import

Users can be imported using their name, e.g.

```
$ terraform import artifactory_user.test-user myusername
```
