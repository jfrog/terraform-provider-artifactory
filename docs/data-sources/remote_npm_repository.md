---
subcategory: "Remote Repositories"
---
# Artifactory Remote Npm Repository Data Source

Retrieves a remote Npm repository.

## Example Usage

```hcl
data "artifactory_remote_npm_repository" "remote-npm" {
  key = "remote-npm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.