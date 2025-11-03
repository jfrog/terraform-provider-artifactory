---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Hex Repository Data Source

Retrieves a virtual Hex repository.

## Example Usage

```hcl
data "artifactory_virtual_hex_repository" "virtual-hex" {
  key = "virtual-hex"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the virtual repositories](../resources/virtual.md):

* `hex_primary_keypair_ref` - (Computed) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client. 