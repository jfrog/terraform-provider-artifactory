---
subcategory: "Local Repositories"
---
# Artifactory Local Alpine Repository Resource

Creates a local Alpine repository.

## Example Usage

```hcl
resource "artifactory_keypair" "some-keypairRSA" {
  pair_name         = "some-keypair"
  pair_type         = "RSA"
  alias             = "foo-alias"
  private_key       = file("samples/rsa.priv")
  public_key        = file("samples/rsa.pub")
  lifecycle {
    ignore_changes  = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_local_alpine_repository" "terraform-local-test-alpine-repo-basic" {
  key                 = "terraform-local-test-alpine-repo-basic"
  primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name

  depends_on          = [artifactory_keypair.some-keypairRSA]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):


* `key` - (Required) the identity key of the repo.
* `primary_keypair_ref` - (Optional) The RSA key to be used to sign alpine indices.

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.


The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_alpine_repository.terraform-local-test-alpine-repo-basic terraform-local-test-alpine-repo-basic
```
