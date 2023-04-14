---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Bower Repository Data Source

Retrieves a virtual Bower repository.

## Example Usage

```hcl
data "artifactory_virtual_bower_repository" "virtual-alpine" {
  key = "virtual-alpine"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `external_dependencies_enabled` - (Optional) When set, external dependencies are rewritten. Default value is false.
* `external_dependencies_remote_repo` - (Optional) The remote repository aggregated by this virtual repository in which the external dependency will be cached.
* `external_dependencies_patterns` - (Optional) An Allow List of Ant-style path expressions that specify where external dependencies may be downloaded from. By default, this is set to ** which means that dependencies may be downloaded from any external source.
