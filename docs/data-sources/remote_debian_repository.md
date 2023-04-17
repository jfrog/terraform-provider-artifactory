---
subcategory: "Remote Repositories"
---
# Artifactory Remote Debian Repository Data Source

Retrieves a remote Debian repository.

## Example Usage

```hcl
data "artifactory_remote_debian_repository" "remote-debian" {
  key = "remote-debian"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.