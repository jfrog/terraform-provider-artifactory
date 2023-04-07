---
subcategory: "Federated Repositories"
---
# Artifactory Federated Rpm Repository Resource

Creates a federated Rpm repository.

## Example Usage

```hcl
resource "artifactory_federated_rpm_repository" "terraform-federated-test-rpm-repo" {
  key       = "terraform-federated-test-rpm-repo"

  member {
    url     = "http://tempurl.org/artifactory/terraform-federated-test-rpm-repo"
    enabled = true
  }

  member {
    url     = "http://tempurl2.org/artifactory/terraform-federated-test-rpm-repo-2"
    enabled = true
  }
}
```

## Argument Reference

The following attributes are supported, along with the [list of attributes from the local RPM repository](local_rpm_repository.md):

* `key` - (Required) the identity key of the repo.
* `member` - (Required) The list of Federated members and must contain this repository URL (configured base URL
  `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set.
  Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)
  to set up Federated repositories correctly.
  * `url` - (Required) Full URL to ending with the repository name.
  * `enabled` - (Required) Represents the active state of the federated member. It is supported to change the enabled
    status of my own member. The config will be updated on the other federated members automatically.
* `cleanup_on_delete` - (Optional) Delete all federated members on `terraform destroy` if set to `true`. Default is `false`. This attribute is added to match Terrform logic, so all the resources, created by the provider, must be removed on cleanup. Artifactory's behavior for the federated repositories is different, all the federated repositories stay after the user deletes the initial federated repository. **Caution**: if set to `true` all the repositories in the federation will be deleted, including repositories on other Artifactory instances in the "Circle of trust". This operation can not be reversed.


## Import

Federated repositories can be imported using their name, e.g.
```
$ terraform import artifactory_federated_rpm_repository.terraform-federated-test-rpm-repo terraform-federated-test-rpm-repo
```
