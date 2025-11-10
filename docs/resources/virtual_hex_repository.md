---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Hex Repository Resource

Creates a virtual Hex repository that aggregates local and remote Hex repositories.

Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Hex+Registry#HexRegistry-VirtualRepositories).

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

resource "artifactory_local_hex_repository" "local-hex" {
  key                     = "local-hex"
  hex_primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
}

resource "artifactory_remote_hex_repository" "remote-hex" {
  key                     = "remote-hex"
  url                     = "https://repo.hex.pm"
  hex_primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  public_key              = file("samples/rsa.pub")
}

resource "artifactory_virtual_hex_repository" "my-virtual-hex" {
  key                     = "my-virtual-hex"
  hex_primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  repositories           = [
    artifactory_local_hex_repository.local-hex.key,
    artifactory_remote_hex_repository.remote-hex.key
  ]
  description             = "Virtual Hex repository aggregating local and remote"
  notes                   = "Internal repository"
  depends_on              = [
    artifactory_keypair.some-keypairRSA,
    artifactory_local_hex_repository.local-hex,
    artifactory_remote_hex_repository.remote-hex
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `hex_primary_keypair_ref` - (Required) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `notes` - (Optional)

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.

The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.

## Import

Virtual repositories can be imported using their name, e.g.
```
$ terraform import artifactory_virtual_hex_repository.my-virtual-hex my-virtual-hex
```

