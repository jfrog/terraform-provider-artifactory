---
subcategory: "Remote Repositories"
---

# Artifactory Remote Hex Repository Resource

Creates a remote hex repository in artifactory. Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/hex-repositories).

## Example Usage

```hcl
resource "artifactory_keypair" "key-pair" {
  pair_name   = "key-pair"
  pair_type   = "RSA"
  alias       = "alias"
  private_key = var.private_key
  public_key  = var.public_key
}

resource "artifactory_remote_hex_repository" "terraform-remote-test-hex-repo-basic" {
  key                        = "terraform-remote-test-hex-repo-basic"
  url                        = "https://hex.pm/"
  public_key_ref             = < remote hex registry public key >
  hex_primary_keypair_ref    = artifactory_keypair.key-pair.pair_name
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) the identity key of the repo.
* `url` - (Required) the remote repo URL.
* `public_key_ref` - (Required) Contains the public key used when downloading packages from the Hex remote registry (public, private, or self-hosted Hex server).
* `hex_primary_keypair_ref` - (Required) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_hex_repository.terraform-remote-test-hex-repo-basic terraform-remote-test-hex-repo-basic
``` 