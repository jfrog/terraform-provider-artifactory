# Artifactory Federated Generic Repository Resource

Creates a federated Generic repository

## Example Usage

```hcl
resource "artifactory_federated_generic_repository" "terraform-federated-test-generic-repo" {
  key = "terraform-federated-test-generic-repo"

  member {
    url    = "http://tempurl.org/artifactory/terraform-federated-test-generic-repo"
    enable = true
  }

  member {
    url    = "http://tempurl2.org/artifactory/terraform-federated-test-generic-repo-2"
    enable = true
  }
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/JFROG/Repository+Configuration+JSON#RepositoryConfigurationJSON-FederatedRepository). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `project_key` - (Optional) Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: "DEV" or "PROD"
* `member` - (Required) - The list of Federated members and must contain this repository URL (configured base URL + `/artifactory/` + repo `key`). Note that each of the federated members will need to have a base URL set. Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository) to set up Federated repositories correctly.
    * `url` - (Required) Full URL to ending with the repository name
    * `enabled` - (Required) Represents the active state of the federated member. It is supported to change the enabled status of my own member. The config will be updated on the other federated members automatically.
* `xray_index` - (Optional, Default: false)  Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.

Arguments for federated repository type closely matches the arguments for local generic repository type.
