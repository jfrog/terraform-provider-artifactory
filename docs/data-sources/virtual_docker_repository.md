---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Docker Repository Data Source

Retrieves a virtual Docker repository.

## Example Usage

```hcl
data "artifactory_virtual_docker_repository" "virtual-docker" {
  key = "virtual-docker"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `resolve_docker_tags_by_timestamp` - (Optional) When enabled, in cases where the same Docker tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp. Default values is `false`.
