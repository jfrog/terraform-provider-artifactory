---
subcategory: "Remote Repositories"
---
# Artifactory Remote Bazel Modules Repository Resource

Creates a remote Bazel Modules repository that proxies and caches modules from a Bazel registry, such as the [Bazel Central Registry (BCR)](https://bcr.bazel.build/). See [Bazel Modules Repositories](https://docs.jfrog.com/artifactory/docs/bazel-modules-repositories).

~> Bazel Modules repositories are supported as **remote** repositories only. Local and virtual Bazel Modules repositories are not supported.

## Example Usage

```hcl
resource "artifactory_remote_bazel_repository" "my-remote-bazelmodules" {
  key         = "my-remote-bazelmodules"
  url         = "https://bcr.bazel.build/"
  description = "Remote Bazel Central Registry proxy"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `url` - (Required) The remote repo URL. For the Bazel Central Registry, use `https://bcr.bazel.build/`.
* `description` - (Optional)
* `notes` - (Optional)

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_bazel_repository.my-remote-bazelmodules my-remote-bazelmodules
```
