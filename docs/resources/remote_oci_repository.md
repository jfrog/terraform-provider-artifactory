---
subcategory: "Remote Repositories"
---
# Artifactory Remote OCI Repository Resource

Creates remote OCI repository resource. 

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/oci-registry).

## Example Usage

```hcl
resource "artifactory_remote_oci_repository" "my-oci-remote" {
  key                            = "my-oci-remote"
  url                            = "https://registry-1.docker.io/"
  external_dependencies_enabled  = true
  external_dependencies_patterns = ["**/registry-1.docker.io/**"]
  enable_token_authentication    = true
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `enable_token_authentication` - (Optional) Enable token (Bearer) based authentication.
* `external_dependencies_enabled` - (Optional) Also known as 'Foreign Layers Caching' on the UI.
* `external_dependencies_patterns` - (Optional) Optional include patterns to match external URLs. Ant-style path expressions are supported (*, **, ?). For example, specifying `**/github.com/**` will only allow downloading foreign layers from github.com host. By default, this is set to `[**]` in the UI, which means that foreign layers may be downloaded from any external hosts. Due to SDKv2 limitations, we can't set the default value for the list. This value `[**]` must be assigned to the attribute manually, if user don't specify any other non-default values. We don't want to make this attribute required, but it must be set to avoid the state drift on update. Note: Artifactory assigns `[**]` on update if HCL doesn't have the attribute set or the list is empty.
* `project_id` (Optional) Use this attribute to enter your GCR, GAR Project Id to limit the scope of this remote repo to a specific project in your third-party registry. When leaving this field blank or unset, remote repositories that support project id will default to their default project as you have set up in your account.

## Import

Remote OCI repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_oci_repository.my-oci-remote my-oci-remote
```
