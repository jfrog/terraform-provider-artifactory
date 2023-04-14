---
subcategory: "Remote Repositories"
---
# Artifactory Remote P2 Repository Data Source

Retrieves a remote P2 repository.

## Example Usage

```hcl
data "artifactory_remote_p2_repository" "remote-p2" {
  key = "remote-p2"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.