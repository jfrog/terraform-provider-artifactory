---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual P2 Repository Data Source

Retrieves a virtual P2 repository.

## Example Usage

```hcl
data "artifactory_virtual_p2_repository" "virtual-p2" {
  key = "virtual-p2"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.
