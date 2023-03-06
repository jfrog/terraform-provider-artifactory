---
subcategory: "Local Repositories"
---

# Artifactory Local PYPI Repository Data Source

Retrieves a local pypi repository.

## Example Usage

```hcl
data "artifactory_local_pypi_repository" "local-test-pypi-repo" {
  key = "local-test-pypi-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
