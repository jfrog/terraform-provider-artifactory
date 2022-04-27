---
subcategory: "Local Repositories"
---
# Artifactory Local Cocoapods Repository Resource

Creates a local Cocoapods repository.

## Example Usage

```hcl
resource "artifactory_local_cocoapods_repository" "terraform-local-test-cocoapods-repo" {
  key = "terraform-local-test-cocoapods-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)



## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_cocoapods_repository.terraform-local-test-cocoapods-repo terraform-local-test-cocoapods-repo
```
