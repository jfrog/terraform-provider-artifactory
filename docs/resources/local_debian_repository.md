---
subcategory: "Local Repositories"
---
# Artifactory Local Debian Repository Resource

Creates a local Debian repository and allows for the creation of a GPG key.

## Example Usage

```hcl
resource "artifactory_keypair" "some-keypairGPG1" {
  pair_name         = "some-keypair${random_id.randid.id}"
  pair_type         = "GPG"
  alias             = "foo-alias1"
  private_key       = file("samples/gpg.priv")
  public_key        = file("samples/gpg.pub")
  lifecycle {
    ignore_changes  = [
      private_key,
      passphrase,
    ]
  }
}
resource "artifactory_keypair" "some-keypairGPG2" {
  pair_name           = "some-keypair4${random_id.randid.id}"
  pair_type           = "GPG"
  alias               = "foo-alias2"
  private_key         = file("samples/gpg.priv")
  public_key          = file("samples/gpg.pub")
  lifecycle {
    ignore_changes    = [
      private_key,
      passphrase,
    ]
  }
}
resource "artifactory_local_debian_repository" "my-debian-repo" {
  key                       = "my-debian-repo"
  primary_keypair_ref       = artifactory_keypair.some-keypairGPG1.pair_name
  secondary_keypair_ref     = artifactory_keypair.some-keypairGPG2.pair_name
  index_compression_formats = ["bz2", "lzma", "xz"]
  trivial_layout            = true
  depends_on                = [artifactory_keypair.some-keypairGPG1, artifactory_keypair.some-keypairGPG2]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `primary_keypair_ref` - (Optional) The primary RSA key to be used to sign packages.
* `secondary_keypair_ref` - (Optional) The secondary RSA key to be used to sign packages.
* `index_compression_formats` - (Optional) The options are Bzip2 (.bz2 extension) (default), LZMA (.lzma extension)
and XZ (.xz extension).
* `trivial_layout` - (Optional) When set, the repository will use the deprecated trivial layout.

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.

The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.



## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_debian_repository.my-debian-repo my-debian-repo
```
