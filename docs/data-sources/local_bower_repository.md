---
subcategory: "Local Repositories"
---

# Artifactory Local Bower Repository Data Source

Retrieves a local Bower repository.

## Example Usage

```hcl
data "artifactory_local_bower_repository" "local-test-bower-repo" {
  key = "local-test-bower-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)
