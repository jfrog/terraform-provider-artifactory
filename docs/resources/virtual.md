---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Repository Common Arguments

The list of arguments, common for the virtual repositories. All these arguments can be used together with the
repository-specific arguments, listed in separate repository-specific documents.  

## Example Usage (generic repository type)

```hcl
resource "artifactory_virtual_generic_repository" "foo-generic" {
  key               = "foo-generic"
  repo_layout_ref   = "simple-default"
  repositories      = []
  description       = "A test virtual repo"
  notes             = "Internal description"
  includes_pattern  = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern  = "com/google/**"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported:

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `project_key` - (Optional) Project key for assigning this repository to. Must be 2 - 20 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash. We don't recommend using this attribute to assign the repository to the project. Use the `repos` attribute in Project provider to manage the list of repositories.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: `DEV`, `PROD`, or one of custom environment.
  Before Artifactory 7.53.1, up to 2 values (`DEV` and `PROD`) are allowed. From 7.53.1 onward, only one value is allowed.
  The attribute should only be used if the repository is already assigned to the existing project. If not, the attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create state drift during the update.
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional) List of artifact patterns to include when evaluating artifact requests in the form of x/y/\*\*/z/\*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/\*).
* `excludes_pattern` - (Optional) List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/*\*/z/\*. By default no artifacts are excluded.
* `repo_layout_ref` - (Optional) Repository layout key for the virtual repository.
* `artifactory_requests_can_retrieve_remote_artifacts` - (Optional, Default: `false`) Whether the virtual repository should search through remote repositories when trying to resolve an artifact requested by another Artifactory instance.
* `default_deployment_repo` - (Optional) Default repository to deploy artifacts.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_generic_repository.foo-generic foo-generic
```
