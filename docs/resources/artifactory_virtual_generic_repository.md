# Artifactory Virtual Generic Repository Resource

Provides an Artifactory virtual repository resource with generic package type. This should be preferred over the original
one-size-fits-all `artifactory_virtual_repository`.

## Example Usage

```hcl
resource "artifactory_virtual_generic_repository" "foo-generic" {
  key          = "foo-generic"
  repo_layout_ref = "simple-default"
  repositories = []
  description = "A test virtual repo"
  notes = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required)
* `repositories` - (Required, but may be empty) The effective list of actual repositories included in this virtual repository.
* `project_key` - (Optional) Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: "DEV" or "PROD"
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional) List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).
* `excludes_pattern` - (Optional) List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.
* `repo_layout_ref` - (Optional)
* `artifactory_requests_can_retrieve_remote_artifacts` - (Optional, Default: false) Whether the virtual repository should search through remote repositories when trying to resolve an artifact requested by another Artifactory instance.
* `default_deployment_repo` - (Optional) Default repository to deploy artifacts.
* `retrieval_cache_period_seconds` - (Optional, Default: 7200) - This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching. Default: 7200 seconds.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_generic_repository.foo foo
```
