---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Pub Repository Data Source

Retrieves a virtual Pub repository.

## Example Usage

```hcl
data "artifactory_virtual_pub_repository" "virtual-pub" {
  key = "virtual-pub"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.
