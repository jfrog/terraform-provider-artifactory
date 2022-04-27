---
subcategory: "Local Repositories"
---
# Artifactory Local Opkg Repository Resource

Creates a local Opkg repository.

## Example Usage

```hcl
resource "artifactory_local_opkg_repository" "terraform-local-test-opkg-repo" {
  key = "terraform-local-test-opkg-repo"
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
$ terraform import artifactory_local_opkg_repository.terraform-local-test-opkg-repo terraform-local-test-opkg-repo
```
