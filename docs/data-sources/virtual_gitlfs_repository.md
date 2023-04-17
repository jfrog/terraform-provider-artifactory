---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Git LFS Repository Data Source

Retrieves a virtual Git LFS repository.

## Example Usage

```hcl
data "artifactory_virtual_gitlfs_repository" "virtual-gitlfs" {
  key = "virtual-gitlfs"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.
