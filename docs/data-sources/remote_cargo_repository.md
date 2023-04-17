---
subcategory: "Remote Repositories"
---
# Artifactory Remote Cargo Repository Data Source

Retrieves a remote Cargo repository.

## Example Usage

```hcl
data "artifactory_remote_cargo_repository" "remote-cargo" {
  key = "remote-cargo"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `anonymous_access` - (Required) Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option. Default value is `false`.
* `enable_sparse_index` - (Optional) Enable internal index support based on Cargo sparse index specifications, instead of the default git index. Default value is `false`.
* `git_registry_url` - (Optional) This is the index url, expected to be a git repository. Default value is `https://github.com/rust-lang/crates.io-index`.
