---
subcategory: "Remote Repositories"
---
# Artifactory Remote Go Repository Data Source

Retrieves a remote Go repository.

## Example Usage

```hcl
data "artifactory_remote_go_repository" "remote-go" {
  key = "remote-go"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is `ARTIFACTORY`.
