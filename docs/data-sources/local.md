---
subcategory: "Local Repositories"
---
# Artifactory Local Repository Data Source

Provides a data source for local repositories.
All local repositories will follow this general format for retrieving the configuration/data for a local repository.

## Example Usage: Generic Local Repository Type

```hcl
#
data "artifactory_local_generic_repository" "local-test-generic-repo-basic" {
  key  = "local-test-generic-repo-basic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) Name of the repository.

## Attribute Reference

In addition to all arguments above, the following attributes are exported for all local repositories:

* `key` - A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - Description of the repository
* `notes`
* `project_key` - Project key for assigning this repository to. Will be 2 - 20 lowercase alphanumeric and
  hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated
  by a dash. We don't recommend using this attribute to assign the repository to the project. Use the `repos` attribute
  in Project provider to manage the list of repositories.
* `project_environments` - Project environment for assigning this repository to. Allow values: `DEV` or `PROD`. The attribute should only be used if the repository is already assigned to the existing project. If not, the
  attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create state drift during
  the update.
* `includes_pattern` - List of artifact patterns to include when evaluating artifact requests in the form of
  x/y/**/z/\*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are
  included (\*\*/*).
* `excludes_pattern` - List of artifact patterns to exclude when evaluating artifact requests, in the form of
  x/y/**/z/*. By default no artifacts are excluded.
* `repo_layout_ref` - Sets the layout that the repository should use for storing and identifying modules. A
  recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and
  calculate metadata accordingly.
* `blacked_out` - (Default: `false`) When set, the repository does not participate in artifact resolution and
  new artifacts cannot be deployed.
* `xray_index` - (Default: `false`) Enable Indexing In Xray. Repository will be indexed with the default
  retention period. You will be able to change it via Xray settings.
* `priority_resolution` - (Default: `false`) Setting repositories with priority will cause metadata to be merged
  only from repositories set with this field
* `property_sets` - List of property set names
* `archive_browsing_enabled` - When set, you may view content such as HTML or Javadoc files directly from
  Artifactory. This may not be safe and therefore requires strict content moderation to prevent malicious users from
  uploading content that may compromise security (e.g., cross-site scripting attacks).
* `download_direct` - When set, download requests to this repository will redirect the client to download the
  artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.

**NOTE:** More attributes are exported for certain repository types. Please see other docs in this folder for specific additional attributes.
