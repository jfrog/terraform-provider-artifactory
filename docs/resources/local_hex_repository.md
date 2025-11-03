---
subcategory: "Local Repositories"
---

# Artifactory Local Hex Repository Resource

Creates a local Hex Repository

```hcl
resource "artifactory_keypair" "key-pair" {
  pair_name   = "key-pair"
  pair_type   = "RSA"
  alias       = "alias"
  private_key = var.private_key
  public_key  = var.public_key
}

resource "artifactory_local_hex_repository" "terraform-local-test-hex-repo-basic" {
  key                        = "terraform-local-test-hex-repo-basic"
  hex_primary_keypair_ref        = artifactory_keypair.key-pair.pair_name
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `hex_primary_keypair_ref` - (Required) Select the RSA key pair to sign and encrypt content for secure communication between Artifactory and the Mix client.


## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_hex_repository.terraform-local-test-hex-repo-basic terraform-local-test-hex-repo-basic
```