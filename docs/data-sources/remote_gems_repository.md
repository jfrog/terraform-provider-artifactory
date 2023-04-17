---
subcategory: "Remote Repositories"
---
# Artifactory Remote Gems Repository Data Source

Retrieves a remote Gems repository.

## Example Usage

```hcl
data "artifactory_remote_gems_repository" "remote-gems" {
  key = "remote-gems"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.