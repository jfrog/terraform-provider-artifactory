---
subcategory: "Local Repositories"
---
# Artifactory Local Hex Repository Resource

Creates a local Hex repository for storing Elixir/Erlang packages.

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

resource "artifactory_local_hex_repository" "my-hex-repo" {
  key                     = "my-hex-repo"
  hex_primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  description             = "Local Hex repository for Elixir packages"
  notes                   = "Internal repository"
  depends_on              = [artifactory_keypair.some-keypairRSA]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `hex_primary_keypair_ref` - (Required) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.
* `description` - (Optional)
* `notes` - (Optional)

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.

The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_hex_repository.my-hex-repo my-hex-repo
```

