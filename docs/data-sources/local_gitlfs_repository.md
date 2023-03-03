---
subcategory: "Local Repositories"
---

# Artifactory Local Gitlfs Repository Data Source

Retrieves a local gitlfs repository.

## Example Usage

```hcl
data "artifactory_local_gitlfs_repository" "local-test-gitlfs-repo" {
  key = "local-test-gitlfs-repo"
}
```

## Attribute Reference

The following attributes are supported along with the [common list of attributes for the local repositories](local.md):

* `key` - the identity key of the repo.
* `description`
* `notes`
