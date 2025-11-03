---
subcategory: "Local Repositories"
---
# Artifactory Local Hex Repository Data Source

Retrieves a local Hex repository.

## Example Usage

```hcl
data "artifactory_local_hex_repository" "local-hex" {
  key = "local-hex"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the local repositories](../resources/local.md):

* `hex_primary_keypair_ref` - (Computed) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client. 