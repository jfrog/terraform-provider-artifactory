---
subcategory: "Remote Repositories"
---
# Artifactory Remote GitLfs Repository Data Source

Retrieves a remote GitLfs repository.

## Example Usage

```hcl
data "artifactory_remote_gitlfs_repository" "remote-gitlfs" {
  key = "remote-gitlfs"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.