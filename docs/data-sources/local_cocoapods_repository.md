---
subcategory: "Local Repositories"
---

# Artifactory Local Cocoapods Repository Data Source

Retrieves a local cocoapods repository.

## Example Usage

```hcl
data "artifactory_local_cocoapods_repository" "local-test-cocoapods-repo" {
  key = "local-test-cocoapods-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
