# Artifactory keypair Resource

Creates an RSA Keypair resource - suitable for signing alpine indices. 
- Currently, only RSA is supported.
- Passphrases are not currently supported, though they exist in the API


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

* `pair_name` - (Required) name of the key pair and the identity of the resource.
* `pair_type` - (Required) RT requires this - presumably for verification purposes.
* `alias` - (Required) Required but for unknown reasons
* `private_key` - (Required)  - duh! This will have it's pem format validated
* `passphrase` - (Optional/Questionable)  - This will be used to decrypt the private key. Validated server side.
* `public_key` - (Required)  - duh! This will have it's pem format validated
* `unavailable` - (Computed) - it's unknown what this does, but, it's returned in the payload and there is no known place to set it in the UI

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.
The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state. 