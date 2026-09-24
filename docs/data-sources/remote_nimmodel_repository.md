---
subcategory: "Remote Repositories"
---

# Artifactory Remote NVIDIA NIM Repository Data Source

Retrieves configuration for a remote NVIDIA NIM repository.

## Example Usage

```hcl
data "artifactory_remote_nimmodel_repository" "example" {
  key = "my-remote-nimmodel"
}
```

## Argument Reference

* `key` - (Required) Repository key.

## Attribute Reference

See the [common attributes for remote repository data sources](remote.md). `package_type` is `nimmodel` and the default layout is `maven-2-default`.

In addition, the following type-specific attribute is exported:

* `enable_token_authentication` - Whether token (Bearer) based authentication is enabled.
