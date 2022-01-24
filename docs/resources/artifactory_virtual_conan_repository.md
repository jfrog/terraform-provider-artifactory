# Artifactory Virtual Conan Repository Resource

Provides an Artifactory virtual repository resource, but with specific conan features. This should be preferred over the original
one-size-fits-all `artifactory_virtual_repository`. 

## Example Usage

```hcl
resource "artifactory_virtual_conan_repository" "foo-conan" {
  key          = "foo-conan"
  repo_layout_ref = "conan-default"
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
* `repositories` - (Required, but may be empty)
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional)
* `excludes_pattern` - (Optional)
* `repo_layout_ref` - (Optional)
* `artifactory_requests_can_retrieve_remote_artifacts` - (Optional)
* `key_pair` - (Optional) - Key pair to use for... well, I'm not sure. Maybe ssh auth to remote repo?
* `default_deployment_repo` - (Optional)
* `virtual_retrieval_cache_period_seconds` - (Optional) - This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching. Default: 7200 seconds.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_conan_repository.foo foo
```
