---
subcategory: "Remote Repositories"
---
# Artifactory Remote VCS Repository Data Source

Retrieves a remote VCS repository.

## Example Usage

```hcl
data "artifactory_remote_vcs_repository" "my-remote-vcs" {
  key = "my-remote-vcs"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):