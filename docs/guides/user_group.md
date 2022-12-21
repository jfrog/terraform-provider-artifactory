---
page_title: "Users and groups management"
---

The way the provider currently provides two resources to manage user-group relationship, `artifactory_user` and `artifactory_group`, which mirrors the Artifactory API, with additional ability to manage groups and users respectively.

However this can create issues such as users not assigned to group or Terraform state drift, if the relationship is specified using `artifactory_group.users_names` attribute or a mixture of both `artifactory_group.users_names` **and** `artifactory_user.groups` attributes.

We recommend managing the user-group relationship using the `artifactory_user` resource only.

## Example

```hcl
resource "artifactory_group" "group-1" {
  name             = "group-1"
  description      = "test group 1"
  external_id      = "00628948-b509-4362-aa73-380c4dbd2a44"
  admin_privileges = false
}

resource "artifactory_group" "group-2" {
  name             = "group-2"
  description      = "test group 2"
  admin_privileges = true
}

resource "artifactory_user" "user-1" {
  name     = "user-1"
  email    = "test-user-1@yourcomany.com"
  password = "my super secret password"

  groups = [
    "readers",
    artifactory_group.group-1.name,
    artifactory_group.group-2.name,
  ]
}

resource "artifactory_user" "user-2" {
  name     = "user-2"
  email    = "test-user-2@yourcomany.com"
  password = "my super secret password"

  groups = [
    "readers",
    artifactory_group.group-2.name,
  ]
}
```
