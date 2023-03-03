---
subcategory: "Local Repositories"
---

# Artifactory Local Terraform Backend Repository Data Source

Retrieves a local terraformbackend repository.

## Example Usage

```hcl
data "artifactory_local_terraformbackend_repository" "local-test-terraformbackend-repo" {
  key = "local-test-terraformbackend-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
