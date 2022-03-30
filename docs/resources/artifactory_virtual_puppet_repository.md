# Artifactory Virtual Puppet Repository Resource

Provides an Artifactory virtual repository resource with specific puppet features. 

## Example Usage

```hcl
resource "artifactory_virtual_puppet_repository" "foo-puppet" {
  key          = "foo-puppet"
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

Arguments for Puppet repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_puppet_repository.foo foo
```
