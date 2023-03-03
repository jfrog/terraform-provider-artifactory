---
subcategory: "Local Repositories"
---

# Artifactory Local Helm Repository Data Source

Retrieves a local helm repository.

## Example Usage

```hcl
data "artifactory_local_helm_repository" "local-test-helm-repo" {
  key = "local-test-helm-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
