---
subcategory: "Local Repositories"
---

# Artifactory Local Conan Repository Data Source

Retrieves a local conan repository.

## Example Usage

```hcl
data "artifactory_local_conan_repository" "local-test-conan-repo" {
  key = "local-test-conan-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
