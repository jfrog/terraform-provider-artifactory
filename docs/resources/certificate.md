---
subcategory: "Security"
---
# Artifactory Certificate Resource

Provides an Artifactory certificate resource. This can be used to create and manage Artifactory certificates which can be used as client authentication against remote repositories.

## Example Usage

```hcl
# Create a new Artifactory certificate called my-cert
resource "artifactory_certificate" "my-cert" {
  alias   = "my-cert"
  content = "${file("/path/to/bundle.pem")}"
}

# This can then be used by a remote repository
resource "artifactory_remote_maven_repository" "my-remote" {
  // more code
  client_tls_certificate = "${artifactory_certificate.my-cert.alias}"
  // more code
}
```

## Argument Reference

The following arguments are supported:

* `alias` - (Required) Name of certificate.
* `content` - (Required) PEM-encoded client certificate and private key.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `fingerprint` - SHA256 fingerprint of the certificate.
* `issued_by` - Name of the certificate authority that issued the certificate.
* `issued_on` - The time & date when the certificate is valid from.
* `issued_to` - Name of whom the certificate has been issued to.
* `valid_until` - The time & date when the certificate expires.

## Import

Certificates can be imported using their alias, e.g.

```
$ terraform import artifactory_certificate.my-cert my-cert
```
