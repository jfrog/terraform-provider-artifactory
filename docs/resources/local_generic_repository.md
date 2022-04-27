---
subcategory: "Local Repositories"
---
# Artifactory Local Generic Repository Resource

Creates a local Generic repository.

## Example Usage

```hcl
resource "artifactory_local_generic_repository" "terraform-local-test-generic-repo" {
  key = "terraform-local-test-generic-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. 
It cannot begin with a number or contain spaces or special characters.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_generic_repository.terraform-local-test-generic-repo terraform-local-test-generic-repo
```
