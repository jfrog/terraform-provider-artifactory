# Artifactory Keypair Resource

Creates a Keypair resource.

~> **Note:** Presently, only **RSA** keys are supported which are suitable for signing Alpine Linux indices. Passphrases are not currently supported, though they exist in the API for GPG keys. GPG support is stubbed-out in the provider, but not fully implemented.

## Example Usage

```hcl
resource "artifactory_keypair" "my_keypair" {
  pair_type   = "RSA"
  pair_name   = "some-keypair6543461672124900137"
  alias       = "foo-alias6543461672124900137"

  private_key = file("samples/rsa.pem")
  public_key  = file("samples/rsa.pub")
  # passphrase = ""

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

* `pair_name` - (Required) Name of the keypair and the identity of the resource.

* `pair_type` - (Required) The type of key. Allowed values are: `RSA` or `GPG`.

* `alias` - (Required) Appears with the keypairs in the Admin UI for _signing keys_.

* `private_key` - (Required) The private key portion of the RSA/GPG keypair, as a string.

* `passphrase` - (Optional) - The passphrase for the GPG key. Validated server side. (Not currently implemented until GPG support is added.)

* `public_key` - (Required) The public key portion of the RSA/GPG keypair, as a string.

~> **Note:** The Artifactory API call _Get Key Pair_ doesn't return `private_key` or `passphrase`, but consumes these keys in the `POST` call. The meta-argument `lifecycle` is used here to ensure the provider ignores the changes for these two keys in the Terraform state.
