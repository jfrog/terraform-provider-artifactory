---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual NPM Repository Data Source

Retrieves a virtual NPM repository.

## Example Usage

```hcl
data "artifactory_virtual_npm_repository" "virtual-npm" {
  key = "virtual-npm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `force_nuget_authentication` - (Optional) If set, user authentication is required when accessing the repository. An anonymous request will display an HTTP 401 error. This is also enforced when aggregated repositories support anonymous requests. Default is `false`.
