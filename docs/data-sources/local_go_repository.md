---
subcategory: "Local Repositories"
---

# Artifactory Local Go Repository Data Source

Retrieves a local go repository.

## Example Usage

```hcl
data "artifactory_local_go_repository" "local-test-go-repo" {
  key = "local-test-go-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
