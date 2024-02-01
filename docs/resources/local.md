---
subcategory: "Local Repositories"
---
# Artifactory Local Repository Common Arguments

The list of arguments, common for the local repositories. All these arguments can be used together with the
repository-specific arguments, listed in separate repository-specific documents.   


## Example Usage (generic repository type)

```hcl
resource "artifactory_local_generic_repository" "terraform-local-test-generic-repo" {
  key = "terraform-local-test-generic-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported:

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `project_key` - (Optional) Project key for assigning this repository to. Must be 2 - 20 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash. We don't recommend using this attribute to assign the repository to the project. Use the `repos` attribute in Project provider to manage the list of repositories.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: `DEV`, `PROD`, or one of custom environment.
  Before Artifactory 7.53.1, up to 2 values (`DEV` and `PROD`) are allowed. From 7.53.1 onward, only one value is allowed.
  The attribute should only be used if the repository is already assigned to the existing project. If not, the attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create state drift during the update.
* `includes_pattern` - (Optional) List of artifact patterns to include when evaluating artifact requests in the form
of x/y/**/z/\*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (\*\*/*).
* `excludes_pattern` - (Optional) List of artifact patterns to exclude when evaluating artifact requests, in the form
of x/y/**/z/*. By default no artifacts are excluded.
* `repo_layout_ref` - (Optional) Sets the layout that the repository should use for storing and identifying modules.
  A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.
* `blacked_out` - (Optional, Default: `false`) When set, the repository does not participate in artifact resolution and
new artifacts cannot be deployed.
* `xray_index` - (Optional, Default: `false`) Enable Indexing In Xray. Repository will be indexed with the default
retention period. You will be able to change it via Xray settings.
* `priority_resolution` - (Optional, Default: `false`) Setting repositories with priority will cause metadata to be
merged only from repositories set with this field
* `property_sets` - (Optional) List of property set name
* `archive_browsing_enabled` - (Optional) When set, you may view content such as HTML or Javadoc files directly from
Artifactory. This may not be safe and therefore requires strict content moderation to prevent malicious users from
uploading content that may compromise security (e.g., cross-site scripting attacks).
* `download_direct` - (Optional) When set, download requests to this repository will redirect the client to download
the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.
* `cdn_redirect` - (Optional) When set, download requests to this repository will redirect the client to download
the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only.
