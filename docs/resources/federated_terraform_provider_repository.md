---
subcategory: "Federated Repositories"
---
# Artifactory Federated Terraform Provider Repository Resource

Creates a federated Terraform Provider repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Terraform+Repositories).

The Terraform Registry protocol requires a configured GPG key pair. Use `primary_keypair_ref` to attach a signing key.

## Example Usage

```hcl
resource "artifactory_keypair" "terraform-signing-key" {
  pair_name   = "terraform-signing-key"
  pair_type   = "GPG"
  alias       = "terraform-provider-signing"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")

  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_federated_terraform_provider_repository" "terraform-federated-test-terraform_provider-repo" {
  key                 = "terraform-federated-test-terraform-provider-repo"
  primary_keypair_ref = artifactory_keypair.terraform-signing-key.pair_name

  member {
    url     = "http://tempurl.org/artifactory/terraform-federated-test-terraform_provider-repo"
    enabled = true
  }

  member {
    url     = "http://tempurl2.org/artifactory/terraform-federated-test-terraform_provider-repo-2"
    enabled = true
  }

  depends_on = [artifactory_keypair.terraform-signing-key]
}
```

## Argument Reference

The following attributes are supported, along with the [list of attributes from the local Terraform Provider repository](local_terraform_provider_repository.md):

* `key` - (Required) the identity key of the repo.
* `primary_keypair_ref` - (Optional) The primary GPG key to be used to sign packages. Default is empty.
* `member` - (Required) The list of Federated members and must contain this repository URL (configured base URL
  `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set.
  Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)
  to set up Federated repositories correctly.
  * `url` - (Required) Full URL to ending with the repository name.
  * `enabled` - (Required) Represents the active state of the federated member. It is supported to change the enabled
    status of my own member. The config will be updated on the other federated members automatically.
  * `access_token` - (Optional) Admin access token for this member Artifactory instance. Used in conjunction with `cleanup_on_delete` attribute when Access Federation for access tokens is not enabled.
* `cleanup_on_delete` - (Optional) Delete all federated members on `terraform destroy` if set to `true`. Default is `false`. This attribute is added to match Terrform logic, so all the resources, created by the provider, must be removed on cleanup. Artifactory's behavior for the federated repositories is different, all the federated repositories stay after the user deletes the initial federated repository. **Caution**: if set to `true` all the repositories in the federation will be deleted, including repositories on other Artifactory instances in the "Circle of trust". This operation can not be reversed.
* `proxy` - (Optional) Proxy key from Artifactory Proxies settings. Default is empty field. Can't be set if `disable_proxy = true`.
* `disable_proxy` - (Optional, Default: `false`) When set to `true`, the proxy is disabled, and not returned in the API response body. If there is a default proxy set for the Artifactory instance, it will be ignored, too.

Artifactory REST API call Get Key Pair doesn't return keys `private_key` and `passphrase`, but consumes these keys in the POST call.

The meta-argument `lifecycle` used here to make Provider ignore the changes for these two keys in the Terraform state.

## Import

Federated repositories can be imported using their name, e.g.
```
$ terraform import artifactory_federated_terraform_provider_repository.terraform-federated-test-terraform_provider-repo terraform-federated-test-terraform-provider-repo
```
