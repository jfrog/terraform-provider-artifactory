---
subcategory: "Local Repositories"
---

# Artifactory Local Vagrant Repository Data Source

Retrieves a local vagrant repository.

## Example Usage

```hcl
data "artifactory_local_vagrant_repository" "local-test-vagrant-repo" {
  key = "local-test-vagrant-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
