---
subcategory: "Security"
---
# Artifactory Permission Target Resource

Provides an Artifactory permission target resource. This can be used to create and manage Artifactory permission targets.

## Example Usage

```hcl
# Create a new Artifactory permission target called testpermission
resource "artifactory_permission_target" "test-perm" {
  name = "test-perm"

  repo {
    includes_pattern = ["foo/**"]
    excludes_pattern = ["bar/**"]
    repositories     = ["example-repo-local"]

    actions {
      users {
        name        = "anonymous"
        permissions = ["read", "write"]
      }

      groups {
        name        = "readers"
        permissions = ["read"]
      }
    }
  }

  build {
    includes_pattern = ["**"]
    repositories     = ["artifactory-build-info"]

    actions {
      users {
        name        = "anonymous"
        permissions = ["read", "write"]
      }
    }
  }

  release_bundle {
    includes_pattern = ["**"]
    repositories     = ["release-bundles"]

    actions {
      users {
        name         = "anonymous"
        permissions  = ["read"]
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of permission.
* `repo` - (Optional) Repository permission configuration.
    * `includes_pattern` - (Optional) Pattern of artifacts to include.
    * `excludes_pattern` - (Optional) Pattern of artifacts to exclude.
    * `repositories` - (Required) List of repositories this permission target is applicable for. You can specify the name `ANY` in the repositories section in order to apply to all repositories, `ANY REMOTE` for all remote repositories and `ANY LOCAL` for all local repositories. The default value will be `[]` if nothing is specified.
    * `actions` -
        * `users` - (Optional) Users this permission target applies for.
        * `groups` - (Optional) Groups this permission applies for.
* `build` - (Optional) As for repo but for artifactory-build-info permissions.
* `release_bundle` - (Optional) As for repo for for release-bundles permissions.

## Permissions

The provider supports the following `permission` enums:

* `read`
* `write`
* `annotate`
* `delete`
* `manage`
* `managedXrayMeta`
* `distribute`

The values can be mapped to the permissions from the official [documentation](https://www.jfrog.com/confluence/display/JFROG/Permissions):

* `read` - matches `Read` permissions.
* `write` - matches ` Deploy / Cache / Create` permissions.
* `annotate` - matches `Annotate` permissions.
* `delete` - matches `Delete / Overwrite` permissions.
* `manage` - matches `Manage` permissions.
* `managedXrayMeta` - matches `Manage Xray Metadata` permissions.
* `distribute` - matches `Distribute` permissions.

## Import

Permission targets can be imported using their name, e.g.

```
$ terraform import artifactory_permission_target.terraform-test-permission mypermission
```
