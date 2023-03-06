---
subcategory: "Local Repositories"
---

# Artifactory Local Cargo Repository Data Source

Retrieves a local cargo repository.

## Example Usage

```hcl
data "artifactory_local_cargo_repository" "local-test-cargo-repo-basic" {
  key = "local-test-cargo-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](local.md):

* `anonymous_access` - Cargo client does not send credentials when performing download and search for crates.
  Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous
  access option. Default value is `false`.
* `enable_sparse_index` - Enable internal index support based on Cargo sparse index specifications, instead
  of the default git index. Default value is `false`.
