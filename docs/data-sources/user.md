# Artifactory User Data Source

Provides an Artifactory user data source. This can be used to read the configuration of users in artifactory.

## Example Usage

```hcl
#
data "artifactory_user" "user1" {
  name  = "user1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the user.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `email` - Email for user.
* `admin` - When enabled, this user is an administrator with all the ensuing privileges. Default value is `false`.
* `profile_updatable` - When set, this user can update his profile details (except for the password. Only an administrator can update the password). Default value is `true`.
* `disable_ui_access` - When set, this user can only access Artifactory through the REST API. This option cannot be set if the user has Admin privileges. Default value is `true`.
* `internal_password_disabled` - When set, disables the fallback of using an internal password when external authentication (such as LDAP) is enabled.
* `groups` - List of groups this user is a part of.
