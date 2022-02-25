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

Arguments for Conan repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_conan_repository.foo foo
```
