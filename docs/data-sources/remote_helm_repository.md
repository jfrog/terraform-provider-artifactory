---
subcategory: "Remote Repositories"
---
# Artifactory Remote Helm Repository Data Source

Retrieves a remote Helm repository.

## Example Usage

```hcl
data "artifactory_remote_helm_repository" "remote-helm" {
  key = "remote-helm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `helm_charts_base_url` - (Optional) No documentation is available. Hopefully you know what this means.
* `external_dependencies_enabled` - (Optional) When set, external dependencies are rewritten. `External Dependency Rewrite` in the UI.
* `external_dependencies_patterns` - (Optional) An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response. By default, this is set to `[**]` in the UI, which means that remote modules may be downloaded from any external VCS source. Due to SDKv2 limitations, we can't set the default value for the list. This value `[**]` must be assigned to the attribute manually, if user don't specify any other non-default values. We don't want to make this attribute required, but it must be set to avoid the state drift on update. Note: Artifactory assigns `[**]` on update if HCL doesn't have the attribute set or the list is empty.