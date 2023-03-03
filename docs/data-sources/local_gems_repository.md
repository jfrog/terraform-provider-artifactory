---
subcategory: "Local Repositories"
---

# Artifactory Local Gems Repository Data Source

Retrieves a local gems repository.

## Example Usage

```hcl
data "artifactory_local_gems_repository" "local-test-gems-repo" {
  key = "local-test-gems-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
