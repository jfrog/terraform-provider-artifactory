---
subcategory: "Local Repositories"
---

# Artifactory Local Swift Repository Data Source

Retrieves a local swift repository.

## Example Usage

```hcl
data "artifactory_local_swift_repository" "local-test-swift-repo" {
  key = "local-test-swift-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
