---
subcategory: "Federated Repositories"
---
# Artifactory Federated SBT Repository Data Source

Retrieves a federated SBT repository.

## Example Usage

```hcl
data "artifactory_federated_sbt_repository" "federated-test-sbt-repo" {
  key = "federated-test-sbt-repo"
}
```

## Argument Reference

* `key` - (Required) the identity key of the repo.

## Attribute Reference
The following attributes are supported, along with the [list of attributes from the local SBT repository](local_sbt_repository.md):

* `member` - The list of Federated members and must contain this repository URL (configured base URL
  `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set.
  Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)
  to set up Federated repositories correctly.
  * `url` - Full URL to ending with the repository name.
  * `enabled` - Represents the active state of the federated member. It is supported to change the enabled
    status of my own member. The config will be updated on the other federated members automatically.
