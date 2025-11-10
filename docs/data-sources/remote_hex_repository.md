---
subcategory: "Remote Repositories"
---
# Artifactory Remote Hex Repository Data Source

Retrieves a remote Hex repository.

## Example Usage

```hcl
data "artifactory_remote_hex_repository" "remote-hex" {
  key = "remote-hex"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `hex_primary_keypair_ref` - Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.
* `public_key` - Contains the public key used when downloading packages from the Hex remote registry (public, private, or self-hosted Hex server).
* `url` - The remote repo URL. For the official Hex registry, use `https://repo.hex.pm`.

