---
subcategory: "User"
---
# Artifactory Anonymous User Resource

Provides an Artifactory anonymous user resource. This can be used to import Artifactory 'anonymous' user for some use cases where this is useful.

This resource is not intended for managing the 'anonymous' user in Artifactory. Use the `resource_artifactory_user` resource instead.

!> Anonymous user cannot be created from scratch, nor updated/deleted once imported into Terraform state.

## Example Usage

```hcl
# Define a new Artifactory 'anonymous' user for import
resource "artifactory_anonymous_user" "anonymous" {
}
```

## Argument Reference

The following argument is supported:

* `name` - (Optional) Username for anonymous user. This is only for ensuring resource schema is valid for Terraform. This is not meant to be set or updated in the HCL.

## Import

Anonymous user can be imported using their name, e.g.

```
$ terraform import artifactory_anonymous_user.anonymous-user anonymous
```
