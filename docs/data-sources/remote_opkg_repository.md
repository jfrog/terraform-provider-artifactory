---
subcategory: "Remote Repositories"
---
# Artifactory Remote Opkg Repository Data Source

Retrieves a remote Opkg repository.

## Example Usage

```hcl
data "artifactory_remote_opkg_repository" "remote-opkg" {
  key = "remote-opkg"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.