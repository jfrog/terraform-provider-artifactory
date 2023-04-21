---
subcategory: "Security"
---
# Artifactory Trusted Key Resource

Provides an Artifactory Trusted Key resource. This can be used to create and manage Artifactory Trusted Keys.

## Example Usage

```hcl
resource "artifactory_trusted_key" "my-key" {
  alias      = "my-key"
  public_key = file("samples/rsa.pub")
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) Will be used as a identifier when uploading/retrieving the public key via REST API.
* `public_key` - (Required) The Public key to add as trusted distribution GPG key.

The following additional attributes are exported:

* `key_id` - Returns the key id by which this key is referenced in Artifactory
* `fingerprint` - Returns the computed key fingerprint
* `issued_on` - Returns the date/time when this GPG key was created
* `issued_by` - Returns the name and eMail address of issuer
* `valid_until` - Returns the date/time until this GPG key is valid for

## Import

Currently not supported