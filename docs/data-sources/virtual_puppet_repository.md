---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Puppet Repository Data Source

Retrieves a virtual Puppet repository.

## Example Usage

```hcl
data "artifactory_virtual_puppet_repository" "virtual-puppet" {
  key = "virtual-puppet"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.
