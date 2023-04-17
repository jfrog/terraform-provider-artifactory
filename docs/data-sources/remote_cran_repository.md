---
subcategory: "Remote Repositories"
---
# Artifactory Remote CRAN Repository Data Source

Retrieves a remote CRAN repository.

## Example Usage

```hcl
data "artifactory_remote_cran_repository" "remote-cran" {
  key = "remote-cran"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the remote repositories](../resources/remote.md) is supported.