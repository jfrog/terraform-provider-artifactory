---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual PHP Composer Repository Data Source

Retrieves a virtual PHP Composer repository.

## Example Usage

```hcl
data "artifactory_virtual_composer_repository" "virtual-composer" {
  key = "virtual-composer"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.