# Artifactory Virtual Bower Repository Resource

Provides an Artifactory virtual repository resource with specific bower features. 

## Example Usage

```hcl
resource "artifactory_virtual_bower_repository" "foo-bower" {
  key          = "foo-bower"
  repositories = []
  description = "A test virtual repo"
  notes = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
  external_dependencies_enabled = false
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required)
* `repositories` - (Required, but may be empty)
* `description` - (Optional)
* `notes` - (Optional)
* `external_dependencies_enabled` - (Optional) When set, external dependencies are rewritten. Default value is false.
* `external_dependencies_remote_repo` - (Optional) The remote repository aggregated by this virtual repository in which the external dependency will be cached.
* `external_dependencies_patterns` - (Optional) An Allow List of Ant-style path expressions that specify where external dependencies may be downloaded from. By default, this is set to ** which means that dependencies may be downloaded from any external source.

Arguments for Bower repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_bower_repository.foo foo
```
