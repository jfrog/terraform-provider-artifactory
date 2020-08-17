---
layout: "artifactory"
page_title: "Artifactory: artifactory_permission_target"
sidebar_current: "docs-artifactory-resource-permission-target"
description: |-
  Provides an permission target resource.
---

# artifactory_permission_target

**Requires Artifactory >= 6.6.0. If using a lower version see [here]()**

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
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of permission
* `repo` - (Optional) Repository permission configuration
    * `includes_pattern` - (Optional) Pattern of artifacts to include
    * `excludes_pattern` - (Optional) Pattern of artifacts to exclude
    * `repositories` - (Optional) List of repositories this permission target is applicable for
    * `actions` -
        * `users` - (Optional) Users this permission target applies for. 
        * `groups` - (Optional) Groups this permission applies for. 
* `build` - (Optional) As for repo but for artifactory-build-info permssions.

## Import

Permission targets can be imported using their name, e.g.

```
$ terraform import artifactory_permission_target.terraform-test-permission mypermission
```
