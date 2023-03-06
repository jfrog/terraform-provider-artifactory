---
subcategory: "Local Repositories"
---

# Artifactory Local Cran Repository Data Source

Retrieves a local cran repository.

## Example Usage

```hcl
data "artifactory_local_cran_repository" "local-test-cran-repo" {
  key = "local-test-cran-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
