---
subcategory: "Security"
---
# Artifactory Distribution Public Key Resource

Provides an Artifactory Distribution Public Key resource. This can be used to create and manage Artifactory Distribution Public Keys.

See [API description](https://jfrog.com/help/r/jfrog-rest-apis/set-distributionpublic-gpg-key) in the Artifactory documentation for more details. Also the [UI documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/managing-webstart-and-jar-signing) has further details on where to find these keys in Artifactory.


## Example Usage

```hcl
resource "artifactory_distribution_public_key" "my-key" {
  alias      = "my-key"
  public_key = file("samples/rsa.pub")
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) Will be used as an identifier when uploading/retrieving the public key via REST API.
* `public_key` - (Required) The Public key to add as a trusted distribution GPG key.

The following additional attributes are exported:

* `key_id` - Returns the key id by which this key is referenced in Artifactory
* `fingerprint` - Returns the computed key fingerprint
* `issued_on` - Returns the date/time when this GPG key was created
* `issued_by` - Returns the name and eMail address of issuer
* `valid_until` - Returns the date/time when this GPG key expires.

## Import

Distribution Public Key can be imported using the key id, e.g.

```
$ terraform import artifactory_distribution_public_key.my-key keyid
```
