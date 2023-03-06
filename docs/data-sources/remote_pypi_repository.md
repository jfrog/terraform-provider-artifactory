---
subcategory: "Remote Repositories"
---
# Artifactory Remote Pypi Repository Data Source

Retrieves a remote Pypi repository.

## Example Usage

```hcl
data "artifactory_remote_pypi_repository" "my-remote-pypi" {
  key = "my-remote-pypi"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](remote.md):