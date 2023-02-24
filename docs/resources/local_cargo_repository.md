---
subcategory: "Local Repositories"
---
# Artifactory Local Cargo Repository Resource

Creates a local Cargo repository.

## Example Usage

```hcl
resource "artifactory_local_cargo_repository" "terraform-local-test-cargo-repo-basic" {
  key                 = "terraform-local-test-cargo-repo-basic"
  anonymous_access    = false
  enable_sparse_index = true
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `anonymous_access` - (Optional) Cargo client does not send credentials when performing download and search for crates. 
Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is `false`.
* `enable_sparse_index` - (Optional) Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is `false`.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_cargo_repository.terraform-local-test-cargo-repo-basic terraform-local-test-cargo-repo-basic
```
