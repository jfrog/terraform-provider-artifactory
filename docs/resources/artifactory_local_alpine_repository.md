# Artifactory Local Alpine Repository Resource

Creates a local Alpine repository and allows for the creation of a 

## Example Usage

```hcl
resource "artifactory_keypair" "some-keypairRSA" {
  pair_name   = "some-keypair"
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

resource "artifactory_local_alpine_repository" "terraform-local-test-alpine-repo-basic" {
  key                 = "terraform-local-test-alpine-repo-basic"
  primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  depends_on = [artifactory_keypair.some-keypairRSA]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `primary_keypair_ref` - (Optional) - The RSA key to be used to sign alpine indecies