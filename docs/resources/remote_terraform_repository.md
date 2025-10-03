---
subcategory: "Remote Repositories"
---
# Artifactory Remote Repository Resource

Creates a remote Terraform repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Terraform+Repositories).

## Example Usage

```hcl
resource "artifactory_remote_terraform_repository" "terraform-remote" {
  key                     = "terraform-remote"
  url                     = "https://github.com/"
  terraform_registry_url  = "https://registry.terraform.io"
  terraform_providers_url = "https://releases.hashicorp.com"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The base URL of the Module storage API.
* `terraform_registry_url` - (Optional) The base URL of the registry API. 
  When using Smart Remote Repositories, set the URL to `<base_Artifactory_URL>/api/terraform/repokey`.
* `terraform_providers_url` - (Optional) The base URL of the Provider's storage API.
  When using Smart remote repositories, set the URL to `<base_Artifactory_URL>/api/terraform/repokey/providers`.

## Important Notes

### Bypass HEAD Requests Requirement

For Terraform remote repositories using the following registry URLs, the `bypass_head_requests` parameter **must** be set to `true`:

- `https://registry.terraform.io` (all Artifactory versions)
- `https://registry.opentofu.org` (all Artifactory versions)
- `https://tf.app.wiz.io` (Artifactory 7.122.0 and later only)

Artifactory automatically enforces `bypass_head_requests = true` for these registries, even if you attempt to set it to `false`. This is because these registries do not support HEAD requests, and Artifactory must use GET requests directly to cache artifacts.

**Note**: For `tf.app.wiz.io`, the bypass requirement only applies to Artifactory versions 7.122.0 and later. Earlier versions do not require this setting for the Wiz registry.

**Example with required setting:**
```hcl
resource "artifactory_remote_terraform_repository" "terraform-remote" {
  key                     = "terraform-remote"
  url                     = "https://github.com/"
  terraform_registry_url  = "https://registry.terraform.io"
  terraform_providers_url = "https://releases.hashicorp.com"
  bypass_head_requests    = true  # Required for registry.terraform.io
}
```

If you don't set `bypass_head_requests = true` for these registries (when required for your Artifactory version), you will experience state drift as Artifactory will automatically override the setting to `true`.

**Note**: The `bypass_head_requests` parameter defaults to `false` for most registries. Only the specific registries listed above require it to be set to `true`, and for `tf.app.wiz.io` this requirement only applies to Artifactory 7.122.0 and later.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_terraform_repository.terraform-remote terraform-remote
```
