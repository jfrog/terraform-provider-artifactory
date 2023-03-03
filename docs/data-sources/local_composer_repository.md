---
subcategory: "Local Repositories"
---

# Artifactory Local Composer Repository Data Source

Retrieves a local composer repository.

## Example Usage

```hcl
data "artifactory_local_composer_repository" "local-test-composer-repo" {
  key = "local-test-composer-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
