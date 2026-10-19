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
      source    = "jfrog/artifactory"
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

### Write-only private key (Terraform 1.11+)

To keep the private key out of Terraform state entirely, use the write-only `private_key_wo` attribute (instead of `private_key`) together with an [ephemeral value](https://developer.hashicorp.com/terraform/language/resources/ephemeral). The value is sent to Artifactory but never persisted to state or plan. Since write-only values are not tracked in state, bump `private_key_wo_version` whenever the key changes so the resource is replaced with the new value:

```hcl
ephemeral "aws_secretsmanager_secret_version" "signing_key" {
  secret_id = var.signing_key_secret_id
}

resource "artifactory_keypair" "some-keypair" {
  pair_name              = "some-keypair"
  pair_type              = "RSA"
  alias                  = "some-alias"
  private_key_wo         = ephemeral.aws_secretsmanager_secret_version.signing_key.secret_string
  private_key_wo_version = "1"
  public_key             = file("samples/rsa.pub")
}
```

~> **Note on migrating existing resources:** Switching an existing `artifactory_keypair` resource from `private_key` to `private_key_wo` updates the Terraform state in-place (clearing `private_key` from the current state without replacing the key pair in Artifactory or interrupting signing). However, because Artifactory does not provide an update API for key pairs, the new `private_key_wo` value is only sent to Artifactory when `private_key_wo_version` is subsequently changed (which replaces the resource). Also note that previously saved Terraform state versions / state history will still contain the old `private_key` secret.

## Argument Reference

The following arguments are supported:

* `pair_name` - (Required) A unique identifier for the Key Pair record.
* `pair_type` - (Required) Key Pair type. Supported types - GPG and RSA.
* `alias` - (Required) Will be used as a filename when retrieving the public key via REST API.
* `private_key` - (Optional, Sensitive) - Private key. PEM format will be validated. Must not include extranous spaces or tabs. Conflicts with `private_key_wo`. Either `private_key` or `private_key_wo` must be set.
* `private_key_wo` - (Optional, Sensitive, [Write-only](https://developer.hashicorp.com/terraform/language/resources/ephemeral/write-only)) Write-only equivalent of `private_key`. Private key. PEM format will be validated. Must not include extranous spaces or tabs. The value is **never stored in Terraform state or plan**. Requires Terraform 1.11 or later. Conflicts with `private_key`. Either `private_key` or `private_key_wo` must be set. Because write-only values are not tracked in state, use `private_key_wo_version` to trigger a replacement when the key has changed.
* `private_key_wo_version` - (Optional) A version identifier for `private_key_wo`. Change this value (for example, after rotating the key) to trigger a replacement that re-sends the current `private_key_wo` value to Artifactory. Only meaningful together with `private_key_wo`.
* `passphrase` - (Optional, Sensitive) Passphrase will be used to decrypt the private key. Validated server side.
* `public_key` - (Required) Public key. PEM format will be validated. Must not include extranous spaces or tabs.

Artifactory REST API call 'Get Key Pair' doesn't return attributes `private_key` and `passphrase`, but consumes these keys in the POST call.

## Import

Keypair can be imported using the pair name, e.g.

```
$ terraform import artifactory_keypair.my-keypair my-keypair-name
```
