---
subcategory: "Remote Repositories"
---
# Artifactory Remote Peppet Repository Data Source

Retrieves a remote Peppet repository.

## Example Usage

```hcl
data "artifactory_remote_peppet_repository" "my-remote-peppet" {
  key = "my-remote-peppet"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):