# Artifactory Virtual Alpine Repository Resource

Provides an Artifactory virtual repository resource with specific alpine features. 

## Example Usage

```hcl
resource "artifactory_virtual_alpine_repository" "foo-alpine" {
  key          = "foo-alpine"
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
* `primary_keypair_ref` - (Optional) Primary keypair used to sign artifacts. Default value is empty.
* `retrieval_cache_period_seconds` - (Optional) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching. Default value is 7200.

Arguments for Alpine repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_alpine_repository.foo foo
```
