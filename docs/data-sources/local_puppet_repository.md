---
subcategory: "Local Repositories"
---

# Artifactory Local Puppet Repository Data Source

Retrieves a local puppet repository.

## Example Usage

```hcl
data "artifactory_local_puppet_repository" "local-test-puppet-repo" {
  key = "local-test-puppet-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
