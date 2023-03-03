---
subcategory: "Remote Repositories"
---
# Artifactory Remote Opkg Repository Data Resource

Retrieves a remote Opkg repository.

## Example Usage

```hcl
data "artifactory_remote_opkg_repository" "my-remote-opkg" {
  key = "my-remote-opkg"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):