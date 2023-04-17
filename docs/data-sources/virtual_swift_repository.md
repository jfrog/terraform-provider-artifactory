---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Swift Repository Data Source

Retrieves a virtual Swift repository.

## Example Usage

```hcl
data "artifactory_virtual_swift_repository" "virtual-swift" {
  key = "virtual-swift"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.
