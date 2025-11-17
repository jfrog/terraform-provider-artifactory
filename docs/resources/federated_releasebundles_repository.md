---
subcategory: "Federated Repositories"
---
# Artifactory Federated Release Bundles Repository Resource

Creates a federated Release Bundles repository. This resource should be used only when you require release bundle federation and are using Artifactory Projects.
It is not possible to create repository with this package type in the UI. 

Default Behavior vs. Recommended Use
**Default Behavior**: When you create the first release bundle within a project, Artifactory automatically provisions a local repository named `<project_name>-release-bundles-v2`.

**The Problem**: This local repository must then be manually converted to a federated repository.

**Recommended Solution**: To avoid this manual step (especially with hundred of projects), use this resource to create the federated `releasebundles` repository before creating any release bundles in the project.

This proactive approach ensures all new release bundle assets are written directly to the federated repository.

~>Note: Federation doesn't work correctly for environments, and if member repository will be created by the federation, it will only have one environment - `DEV` disregard on the values on the source repository. As a workaround, user can create repositories on both instances with correct environments, then modify the configuration by adding members to each of the repos on each instance. 

## Example Usage

```hcl
resource "artifactory_federated_releasebundles_repository" "terraform-federated-test-releasebundles-repo" {
  key           = "terraform-federated-test-releasebundles-repo"
  project_key   = project.test.key

  member {
    url     = "https://tempurl.org/artifactory/project-release-bundles-v2"
    enabled = true
  }

  member {
    url     = "https://tempurl2.org/artifactory/project-release-bundles-v2"
    enabled = true
  }
}
```

## Argument Reference

The following attributes are supported, along with the [list of attributes from the local generic repository](local_generic_repository.md):

* `key` - (Required) the identity key of the repo. Repository name for this package type myst chave following format: `<project_name>-release-bundles-v2`. Project myst exist before the repository was created. 
* `project_key` - (Required) Project key for assigning this repository to. Must be 2 - 32 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash. Even thought we don't recommend using this attribute to assign all other repository types to the project, in the scope of this specific package type, it is necessary to have theis attribute.
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

## Import

Federated repositories can be imported using their name, e.g.
```
$ terraform import artifactory_federated_rpm_repository.terraform-federated-test-rpm-repo terraform-federated-test-rpm-repo
```
