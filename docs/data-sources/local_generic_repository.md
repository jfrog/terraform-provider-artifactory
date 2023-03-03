---
subcategory: "Local Repositories"
---

# Artifactory Local Generic Repository Data Source

Retrieves a local generic repository.

## Example Usage

```hcl
data "artifactory_local_generic_repository" "local-test-generic-repo" {
  key = "local-test-generic-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
