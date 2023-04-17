---
subcategory: "Remote Repositories"
---
# Artifactory Remote Generic Repository Data Source

Retrieves a remote Generic repository.

## Example Usage

```hcl
data "artifactory_remote_generic_repository" "remote-generic" {
  key = "remote-generic"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `propagate_query_params` - (Optional, Default: `false`) When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.
