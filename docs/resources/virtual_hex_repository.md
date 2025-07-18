# Artifactory Virtual Hex Repository Resource

Creates a virtual hex repository in artifactory. Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/create-hex-repository).

## Example Usage

```hcl
resource "artifactory_keypair" "key-pair" {
  pair_name   = "key-pair"
  pair_type   = "RSA"
  alias       = "alias"
  private_key = var.private_key
  public_key  = var.public_key
}

resource "artifactory_virtual_hex_repository" "foo" {
  key                    = "foo"
  repositories           = ["local-hex-repo"]
  description            = "A test virtual repo"
  notes                  = "Internal description"
  hex_primary_keypair_ref = artifactory_keypair.key-pair.pair_name
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFROG API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported, along with the [common list of arguments for the virtual repositories](https://github.com/jfrog/terraform-provider-artifactory/blob/master/docs/resources/virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `hex_primary_keypair_ref` - (Required) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_hex_repository.foo foo
``` 