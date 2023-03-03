---
subcategory: "Local Repositories"
---

# Artifactory Local Npm Repository Data Source

Retrieves a local npm repository.

## Example Usage

```hcl
data "artifactory_local_npm_repository" "local-test-npm-repo" {
  key = "local-test-npm-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
