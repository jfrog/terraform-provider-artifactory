---
subcategory: "Local Repositories"
---
# Artifactory Local Terraform Provider Repository Resource

Creates a local Terraform Provider repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Terraform+Repositories).

## Example Usage

```hcl
resource "artifactory_keypair" "terraform-provider-signing-key" {
  pair_name   = "terraform-provider-signing-key"
  pair_type   = "GPG"
  alias       = "terraform-provider-signing-key"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")
}

resource "artifactory_local_terraform_provider_repository" "terraform-local-test-terraform-provider-repo" {
  key                 = "terraform-local-test-terraform-provider-repo"
  primary_keypair_ref = artifactory_keypair.terraform-provider-signing-key.pair_name
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)
* `primary_keypair_ref` - (Optional) The primary GPG key used to sign packages. The Terraform Registry protocol requires the registry to expose this key as `signing_keys.gpg_public_keys` so that `terraform init` can verify downloaded providers; without it, provider installation fails with `Repository '<key>' is missing a signing key`.
* `secondary_keypair_ref` - (Optional) The secondary GPG key used to sign packages.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_terraform_provider_repository.terraform-local-test-terraform-provider-repo terraform-local-test-terraform-provider-repo
```
