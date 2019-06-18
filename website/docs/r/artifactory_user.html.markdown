---
layout: "artifactory"
page_title: "Artifactory: artifactory_user"
sidebar_current: "docs-artifactory-resource-user"
description: |-
  Provides an user resource.
---

# artifactory_user

Provides an Artifactory user resource. This can be used to create and manage Artifactory users.

Note: User passwords are never returned through the API. Since they are never returned they cannot be managed by 
directly through Terraform. However, it is possible to store a "known" state for the password and make changes if it's
updated in Terraform.

## Example Usage

```hcl
# Create a new Artifactory user called terraform
resource "artifactory_user" "test-user" {
  name     = "terraform"
  email    = "test-user@artifactory-terraform.com"
  groups   = ["logged-in-users", "readers"]
  password = "my super secret password"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Username for user
* `email` - (Required) Email for user
* `password` - (Required) Password for the user
* `admin` - (Optional) 
* `profile_updatable` - (Optional) When set, this user can update his profile details (except for the password. Only an administrator can update the password).
* `disable_ui_access` - (Optional) When set, this user can only access Artifactory through the REST API. This option cannot be set if the user has Admin privileges.
* `internal_password_disabled` - (Optional) When set, disables the fallback of using an internal password when external authentication (such as LDAP) is enabled.
* `groups` - (Optional) List of groups this user is a part of

## Import

Users can be imported using their name, e.g.

```
$ terraform import artifactory_user.test-user myusername
```
