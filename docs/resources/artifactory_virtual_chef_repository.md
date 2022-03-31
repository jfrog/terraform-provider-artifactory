# Artifactory Virtual Chef Repository Resource

Provides an Artifactory virtual repository resource, but with specific chef features. This should be preferred over the original
one-size-fits-all `artifactory_virtual_repository`.

## Example Usage

```hcl
resource "artifactory_virtual_chef_repository" "foo-chef" {
  key          = "foo-chef"
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
* `retrieval_cache_period_seconds` - (Optional) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching. Default value is 7200.

Arguments for Chef repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_chef_repository.foo foo
```
