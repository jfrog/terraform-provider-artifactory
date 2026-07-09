---
subcategory: "Remote Repositories"
---

# Artifactory Remote Bazel Modules Repository Data Source

Retrieves configuration for a remote Bazel Modules repository. See [Bazel Modules Repositories](https://docs.jfrog.com/artifactory/docs/bazel-modules-repositories).

## Example Usage

```hcl
data "artifactory_remote_bazel_repository" "example" {
  key = "my-remote-bazelmodules"
}
```

## Argument Reference

* `key` - (Required) Repository key.

## Attribute Reference

See the [common attributes for remote repository data sources](remote.md). `package_type` is `bazelmodules` and the default layout is `simple-default`.
