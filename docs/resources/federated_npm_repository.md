---
subcategory: "Federated Repositories"
---
# Artifactory Federated Npm Repository Resource

Creates a federated Npm repository.

## Example Usage

```hcl
resource "artifactory_federated_npm_repository" "terraform-federated-test-npm-repo" {
  key       = "terraform-federated-test-npm-repo"

  member {
    url     = "http://tempurl.org/artifactory/terraform-federated-test-npm-repo"
    enabled = true
  }

  member {
    url     = "http://tempurl2.org/artifactory/terraform-federated-test-npm-repo-2"
    enabled = true
  }
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/JFROG/Repository+Configuration+JSON#RepositoryConfigurationJSON-FederatedRepository). 
The following arguments are supported, along with the [common list of arguments for the federated repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `member` - (Required) The list of Federated members and must contain this repository URL (configured base URL
  `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set.
  Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)
  to set up Federated repositories correctly.
  * `url` - (Required) Full URL to ending with the repository name.
  * `enabled` - (Required) Represents the active state of the federated member. It is supported to change the enabled
    status of my own member. The config will be updated on the other federated members automatically.



## Import

Federated repositories can be imported using their name, e.g.
```
$ terraform import artifactory_federated_npm_repository.terraform-federated-test-npm-repo terraform-federated-test-npm-repo
```
