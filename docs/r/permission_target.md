---
layout: default
---
# artifactory_permission_target

Provides an Artifactory permission target resource. This can be used to create and manage Artifactory permission targets.

## Example Usage

```hcl
# Create a new Artifactory permission target called testpermission
resource "artifactory_permission_targets" "terraform-test-permission" {
    name          = "testpermission"
    repositories = ["myrepo"]
    users = [
        {
            name = "test_user"
            permissions = [
                "r",
                "w"
            ]
        }
    ]
    
    groups = [
        {
             name        = "readers"
             permissions = ["r"]
        }
    ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of permission
* `includes_pattern` - (Optional) Pattern of artifacts to include
* `excludes_pattern` - (Optional) Pattern of artifacts to exclude
* `repositories` - (Optional) List of repositories this permission target is applicable for
* `users` - (Optional) Users this permission target applies for. 
* `groups` - (Optional) Groups this permission applies for. 

The permissions can be set to a combination of m=admin; d=delete; w=deploy; n=annotate; r=read

## Import

Permission targets can be imported using their name, e.g.

```
$ terraform import artifactory_permission_targets.terraform-test-permission mypermission
```
