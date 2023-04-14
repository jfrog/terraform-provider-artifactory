---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Pypi Repository Data Source

Retrieves a virtual Pypi repository.

## Example Usage

```hcl
data "artifactory_virtual_pypi_repository" "virtual-pypi" {
  key = "virtual-pypi"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.
