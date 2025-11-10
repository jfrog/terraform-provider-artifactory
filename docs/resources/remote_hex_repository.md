---
subcategory: "Remote Repositories"
---
# Artifactory Remote Hex Repository Resource

Creates a remote Hex repository for proxying Elixir/Erlang packages from a remote Hex registry.

Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Hex+Registry).

## Example Usage

```hcl
resource "artifactory_keypair" "some-keypairRSA" {
  pair_name   = "some-keypair${random_id.randid.id}"
  pair_type   = "RSA"
  alias       = "foo-alias"
  private_key = file("samples/rsa.priv")
  public_key  = file("samples/rsa.pub")
  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_remote_hex_repository" "my-remote-hex" {
  key                     = "my-remote-hex"
  url                     = "https://repo.hex.pm"
  hex_primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  public_key              = file("samples/rsa.pub")
  description             = "Remote Hex repository for Elixir packages"
  notes                   = "Internal repository"
  depends_on              = [artifactory_keypair.some-keypairRSA]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `url` - (Required) The remote repo URL. For the official Hex registry, use `https://repo.hex.pm`.
* `hex_primary_keypair_ref` - (Required) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.
* `public_key` - (Required) Contains the public key used when downloading packages from the Hex remote registry (public, private, or self-hosted Hex server).
* `description` - (Optional)
* `notes` - (Optional)

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.

The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_hex_repository.my-remote-hex my-remote-hex
```

