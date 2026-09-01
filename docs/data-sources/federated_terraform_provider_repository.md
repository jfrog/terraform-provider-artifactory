---
subcategory: "Federated Repositories"
---
# Artifactory Federated Terraform Provider Repository Data Source

Retrieves a federated Terraform Provider repository.

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

resource "artifactory_federated_terraform_provider_repository" "federated-test-terraform_provider-repo" {
  key                 = "federated-test-terraform-provider-repo"
  primary_keypair_ref = artifactory_keypair.terraform-signing-key.pair_name

  member {
    url     = "http://tempurl.org/artifactory/federated-test-terraform-provider-repo"
    enabled = true
  }

  depends_on = [artifactory_keypair.terraform-signing-key]
}

data "artifactory_federated_terraform_provider_repository" "federated-test-terraform_provider-repo" {
  key = artifactory_federated_terraform_provider_repository.federated-test-terraform_provider-repo.key
}
```

## Argument Reference

* `key` - (Required) the identity key of the repo.

## Attribute Reference
The following attributes are supported, along with the [list of attributes from the local Terraform Provider repository](local_terraform_provider_repository.md):

* `primary_keypair_ref` - The primary GPG key to be used to sign packages.
* `member` - The list of Federated members and must contain this repository URL (configured base URL
  `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set.
  Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)
  to set up Federated repositories correctly.
  * `url` - Full URL to ending with the repository name.
  * `enabled` - Represents the active state of the federated member. It is supported to change the enabled
    status of my own member. The config will be updated on the other federated members automatically.
* `proxy` - Proxy key from Artifactory Proxies settings.
* `disable_proxy` - When set to `true`, the proxy is disabled, and not returned in the API response body. If there is a default proxy set for the Artifactory instance, it will be ignored, too.
