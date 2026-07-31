---
subcategory: "Remote Repositories"
---
# Artifactory Remote Generic Repository Resource

Creates a remote Generic repository.

## Example Usage

```hcl
resource "artifactory_remote_generic_repository" "my-remote-generic" {
  key = "my-remote-generic"
  url = "http://testartifactory.io/artifactory/example-generic/"
}
```

### Access token (Bearer) authentication

When the remote requires an Artifactory / identity access token, set `enable_token_authentication = true` and put the token in `password` (username may be empty). This matches the UI “Enable Token Authentication” checkbox and avoids putting credentials in `custom_http_headers`.

```hcl
resource "artifactory_remote_generic_repository" "token-auth" {
  key                           = "token-auth-generic"
  url                           = "https://remote.example.com/artifactory/generic-local/"
  password                      = var.remote_access_token
  enable_token_authentication   = true
}
```

### Custom HTTP headers

Use `custom_http_headers` to send up to 5 static headers on every outbound request to the remote URL. A common use case is authenticating to Azure Blob Storage or packagecloud.io.

Each header has a `name`, a `value` (masked in plan output), and an optional `sensitive` flag. When `sensitive = true`, Artifactory encrypts the value server-side. The default is `false` (plaintext). Header values are never read back from Artifactory — the value you configure is preserved in state.

```hcl
resource "artifactory_remote_generic_repository" "azure-blob" {
  key = "azure-blob-generic"
  url = "https://example.blob.core.windows.net/container/"

  custom_http_headers = [
    { name = "x-ms-version", value = "2021-12-02" },
    { name = "x-api-key",    value = "my-secret-token", sensitive = true },
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional) Public description.
* `notes` - (Optional) Internal description.
* `url` - (Required) The remote repo URL.
* `propagate_query_params` - (Optional, Default: `false`) When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.
* `retrieve_sha256_from_server` - (Optional, Default: `false`) When set to `true`, Artifactory retrieves the SHA256 from the remote server if it is not cached in the remote repo.
* `enable_token_authentication` - (Optional, Default: `false`) Enable token (Bearer) based authentication. When set, use an access token as `password` (username may be empty) so Artifactory authenticates to the remote with a Bearer token instead of Basic auth.
* `custom_http_headers` - (Optional) List of up to 5 custom HTTP headers sent on every outbound request to the remote URL. Each entry supports:
  * `name` - (Required) Header name.
  * `value` - (Required) Header value. Masked in Terraform plan output. Stored in state as configured; never read back from Artifactory.
  * `sensitive` - (Optional, Default: `false`) When `true`, Artifactory encrypts the value server-side.


## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_generic_repository.my-remote-generic my-remote-generic
```

Note: `custom_http_headers` values are not read back from Artifactory during import. After importing, run `terraform apply` to push your configured headers.
