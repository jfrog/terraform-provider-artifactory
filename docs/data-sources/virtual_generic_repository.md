---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Generic Repository Data Source

Retrieves a virtual Generic repository.

## Example Usage

```hcl
data "artifactory_virtual_generic_repository" "virtual-generic" {
  key = "virtual-generic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.