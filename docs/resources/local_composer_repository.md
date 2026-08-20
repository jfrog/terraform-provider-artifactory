---
subcategory: "Local Repositories"
---
# Artifactory Local Php-Composer Repository Resource

Creates a local Composer repository.

## Example Usage

```hcl
resource "artifactory_local_composer_repository" "terraform-local-test-composer-repo" {
  key                          = "terraform-local-test-composer-repo"
  enable_composer_v1_indexing  = false
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)
* `enable_composer_v1_indexing` - (Optional, Default: `false`) Enable Composer metadata version 1 indexing.



## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_composer_repository.terraform-local-test-composer-repo terraform-local-test-composer-repo
```
