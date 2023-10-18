---
subcategory: "Security"
---
# Artifactory keypair Resource

RSA key pairs are used to sign and verify the Alpine Linux index files in JFrog Artifactory, while GPG key pairs are
used to sign and validate packages integrity in JFrog Distribution. The JFrog Platform enables you to manage multiple RSA and GPG signing keys through the Keys Management UI and REST API. The JFrog Platform supports managing multiple pairs of GPG signing keys to sign packages for authentication of several package types such as Debian, Opkg, and RPM through the Keys Management UI and REST API.

## Example Usage

```hcl
terraform {
  required_providers {
    artifactory = {
      source    = "registry.terraform.io/jfrog/artifactory"
      version   = "9.7.0"
    }
  }
}

resource "artifactory_keypair" "some-keypair-6543461672124900137" {
  pair_name   = "some-keypair-6543461672124900137"
  pair_type   = "RSA"
  alias       = "some-alias-6543461672124900137"
  private_key = file("samples/rsa.priv")
  public_key  = file("samples/rsa.pub")
  passphrase  = "PASSPHRASE"
}
```

## Argument Reference

The following arguments are supported:

* `pair_name` - (Required) A unique identifier for the Key Pair record.
* `pair_type` - (Required) Key Pair type. Supported types - GPG and RSA.
* `alias` - (Required) Will be used as a filename when retrieving the public key via REST API.
* `private_key` - (Required, Sensitive)  - Private key. PEM format will be validated. Must not include extranous spaces or tabs.
* `passphrase` - (Optional, Sensitive) Passphrase will be used to decrypt the private key. Validated server side.
* `public_key` - (Required) Public key. PEM format will be validated. Must not include extranous spaces or tabs.

Artifactory REST API call 'Get Key Pair' doesn't return attributes `private_key` and `passphrase`, but consumes these keys in the POST call.

## Import

Keypair can be imported using the pair name, e.g.

```
$ terraform import artifactory_keypair.my-keypair my-keypair-name
```
