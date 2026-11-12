---
subcategory: "Federated Repositories"
---
# Artifactory Federated Terraform Module Repository Resource

Creates a federated Generic repository.

## Example Usage

```hcl
resource "artifactory_federated_terraform_module_repository" "terraform-federated-test-terraform_module-repo" {
  key       = "terraform-federated-test-terraform-module-repo"
  member {
    url     = "http://tempurl.org/artifactory/terraform-federated-test-terraform_module-repo"
    enabled = true
  }

  member {
    url     = "http://tempurl2.org/artifactory/terraform-federated-test-terraform_module-repo-2"
    enabled = true
  }
}
```

## Argument Reference

The following attributes are supported, along with the [list of attributes from the local Terraform Module repository](local_terraform_module_repository.md):

* `key` - (Required) the identity key of the repo.
* `member` - (Required) The list of Federated members and must contain this repository URL (configured base URL
  `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set.
  Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)
  to set up Federated repositories correctly.
  * `url` - (Required) Full URL to ending with the repository name.
  * `enabled` - (Required) Represents the active state of the federated member. It is supported to change the enabled
    status of my own member. The config will be updated on the other federated members automatically.
  * `access_token` - (Optional) Admin access token for this member Artifactory instance. Used in conjunction with `cleanup_on_delete` attribute when Access Federation for access tokens is not enabled.
  * `mode` - (Optional) Sync mode of the federated member. The only supported value is `receiver`, which makes the member receive-only (unidirectional sync): it receives changes from the other members but does not send its own. If omitted, the member keeps the default bidirectional sync. If the member was initially created as `receiver`, removing this attribute changes it to bidirectional sync. See [Set Up Unidirectional Sync Using the REST API](https://docs.jfrog.com/artifactory/docs/configure-receiving-only-mode#set-up-unidirectional-sync-using-the-rest-api).
* `cleanup_on_delete` - (Optional) Delete all federated members on `terraform destroy` if set to `true`. Default is `false`. This attribute is added to match Terrform logic, so all the resources, created by the provider, must be removed on cleanup. Artifactory's behavior for the federated repositories is different, all the federated repositories stay after the user deletes the initial federated repository. **Caution**: if set to `true` all the repositories in the federation will be deleted, including repositories on other Artifactory instances in the "Circle of trust". This operation can not be reversed.
* `proxy` - (Optional) Proxy key from Artifactory Proxies settings. Default is empty field. Can't be set if `disable_proxy = true`.
* `disable_proxy` - (Optional, Default: `false`) When set to `true`, the proxy is disabled, and not returned in the API response body. If there is a default proxy set for the Artifactory instance, it will be ignored, too.

## Import

Federated repositories can be imported using their name, e.g.
```
$ terraform import artifactory_federated_terraform_module_repository.terraform-federated-test-terraform_module-repo terraform-federated-test-terraform-module-repo
```
