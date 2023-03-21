---
subcategory: "Federated Repositories"
---
# Artifactory Federated Alpine Repository Data Source

Retrieves a federated Alpine repository.

## Example Usage

```hcl
data "artifactory_federated_alpine_repository" "federated-test-alpine-repo" {
  key = "federated-test-alpine-repo"
}
```
## Argument Reference

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [list of attributes from the local Alpine repository](local_alpine_repository.md):

* `member` - The list of Federated members and must contain this repository URL (configured base URL
  `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set.
   Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)
   to set up Federated repositories correctly.
    * `url` - Full URL to ending with the repository name.
    * `enabled` - Represents the active state of the federated member. It is supported to change the enabled
       status of my own member. The config will be updated on the other federated members automatically.
