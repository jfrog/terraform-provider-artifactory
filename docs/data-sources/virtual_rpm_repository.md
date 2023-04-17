---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Rpm Repository Data Source

Retrieves a virtual Rpm repository.

## Example Usage

```hcl
data "artifactory_virtual_rpm_repository" "virtual-rpm" {
  key = "virtual-rpm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `primary_keypair_ref` - (Optional) The primary GPG key to be used to sign packages.
* `secondary_keypair_ref` - (Optional) The secondary GPG key to be used to sign packages.
