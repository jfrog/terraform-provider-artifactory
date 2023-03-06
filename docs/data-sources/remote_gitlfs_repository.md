---
subcategory: "Remote Repositories"
---
# Artifactory Remote GitLfs Repository Data Source

Retrieves a remote GitLfs repository.

## Example Usage

```hcl
data "artifactory_remote_gitlfs_repository" "my-remote-gitlfs" {
  key = "my-remote-gitlfs"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):