---
subcategory: "Remote Repositories"
---
# Artifactory Remote Conan Repository Data Source

Retrieves a remote Conan repository.

## Example Usage

```hcl
data "artifactory_remote_conan_repository" "remote-conan" {
  key = "remote-conan"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `force_conan_authentication` - (Optional) Force basic authentication credentials in order to use this repository. Default value is `false`.
