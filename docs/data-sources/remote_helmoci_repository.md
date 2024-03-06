---
subcategory: "Remote Repositories"
---
# Artifactory Remote Helm OCI Repository Data Source

Retrieves a remote Helm OCI repository.

## Example Usage

```hcl
data "artifactory_remote_helmoci_repository" "my-helmoci-remote" {
  key = "my-helmoci-remote"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication.
* `external_dependencies_enabled` - (Optional) Also known as 'Foreign Layers Caching' on the UI.
* `external_dependencies_patterns` - (Optional) Optional include patterns to match external URLs. Ant-style path expressions are supported (*, **, ?). For example, specifying `**/github.com/**` will only allow downloading foreign layers from github.com host. Due to SDKv2 limitations, we can't set the default value for the list. This value `[**]` must be assigned to the attribute manually, if user don't specify any other non-default values. We don't want to make this attribute required, but it must be set to avoid the state drift on update. Note: Artifactory assigns `[**]` on update if HCL doesn't have the attribute set or the list is empty.
