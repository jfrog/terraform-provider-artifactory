---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Terraform Repository Data Source

Retrieves a virtual Terraform repository.

## Example Usage

```hcl
data "artifactory_virtual_terraform_repository" "virtual-terraform" {
  key = "virtual-terraform"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The [common list of attributes for the virtual repositories](../resources/virtual.md) is supported.
