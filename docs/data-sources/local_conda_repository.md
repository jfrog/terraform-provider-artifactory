---
subcategory: "Local Repositories"
---

# Artifactory Local Conda Repository Data Source

Retrieves a local conda repository.

## Example Usage

```hcl
data "artifactory_local_conda_repository" "local-test-conda-repo" {
  key = "local-test-conda-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
