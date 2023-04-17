---
subcategory: "Remote Repositories"
---
# Artifactory Remote Conda Repository Data Source

Retrieves a remote Conda repository.

## Example Usage

```hcl
data "artifactory_remote_conda_repository" "remote-conda" {
  key = "remote-conda"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.