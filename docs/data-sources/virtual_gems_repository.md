---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Gems Repository Data Source

Retrieves a virtual Gems repository.

## Example Usage

```hcl
data "artifactory_virtual_gems_repository" "virtual-gems" {
  key = "virtual-gems"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.