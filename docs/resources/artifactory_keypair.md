# Artifactory keypair Resource

Creates an RSA Keypair resource - suitable for signing Alpine Linux indices. 
- Currently, only RSA is supported.
- Passphrases are not currently supported, though they exist in the API.


## Example Usage

```hcl
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.14"
    }
  }
}
resource "artifactory_keypair" "some-keypair6543461672124900137" {
  pair_name   = "some-keypair6543461672124900137"
  pair_type   = "RSA"
  alias       = "foo-alias6543461672124900137"
  private_key = file("samples/rsa.priv")
  public_key  = file("samples/rsa.pub")
  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}
```

## Argument Reference

The following arguments are supported:

* `pair_name` - (Required) A unique identifier for the Key Pair record.
* `pair_type` - (Required) Artifactory requires this - presumably for verification purposes. ?????
* `alias` - (Required) Will be used as a filename when retrieving the public key via REST API.
* `private_key` - (Required, Sensitive)  - Private key. Pem format will be validated.
* `passphrase` - (Optional)  - This will be used to decrypt the private key. Validated server side.
* `public_key` - (Required)  - Public key. Pem format will be validated.
* `unavailable` - (Computed) - This field will be returned in the payload and there is no known place to set it in the UI.

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.
The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.

## Import

Keypair can be imported using their name, e.g.

```
$ terraform import artifactory_keypair.my-keypair my-keypair
```