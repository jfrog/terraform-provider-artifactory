---
subcategory: "Remote Repositories"
---
# Artifactory Remote Bower Repository Data Source

Retrieves a remote Bower repository.

## Example Usage

```hcl
data "artifactory_remote_bower_repository" "remote-bower" {
  key = "remote-bower"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `bower_registry_url` - (Optional) Proxy remote Bower repository. Default value is `https://registry.bower.io`.
