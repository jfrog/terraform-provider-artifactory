# Artifactory Virtual Rpm Repository Resource

Provides an Artifactory virtual repository resource with Rpm package type. This should be preferred over the original one-size-fits-all `artifactory_virtual_repository`.

## Example Usage

```hcl
resource "artifactory_keypair" "primary-keypair" {
  pair_name   = "primary-keypair"
  pair_type   = "GPG"
  alias       = "foo-alias-1"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")

  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_keypair" "secondary-keypair" {
  pair_name   = "secondary-keypair"
  pair_type   = "GPG"
  alias       = "foo-alias-2"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")

  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_virtual_rpm_repository" "foo-rpm-virtual" {
  key                   = "foo-rpm-virtual"

  primary_keypair_ref   = artifactory_keypair.primary-keypair.pair_name
  secondary_keypair_ref = artifactory_keypair.secondary-keypair.pair_name

  depends_on = [
    artifactory_keypair.primary-keypair,
    artifactory_keypair.secondary-keypair,
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/JFROG/Repository+Configuration+JSON#RepositoryConfigurationJSON-VirtualRepository). The following arguments are supported:

* `key` - (Required)
* `primary_keypair_ref` - (Optional) The primary GPG key to be used to sign packages
* `secondary_keypair_ref` - (Optional) The secondary GPG key to be used to sign packages

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.

The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.

Arguments for RPM repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_rpm_repository.foo foo
```
