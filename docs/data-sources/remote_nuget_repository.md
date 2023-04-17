---
subcategory: "Remote Repositories"
---
# Artifactory Remote NuGet Repository Data Source

Retrieves a remote NuGet repository.

## Example Usage

```hcl
data "artifactory_remote_nuget_repository" "remote-nuget" {
  key = "remote-nuget"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `feed_context_path` - (Optional) When proxying a remote NuGet repository, customize feed resource location using this attribute. Default value is `api/v2`.
* `download_context_path` - (Optional) The context path prefix through which NuGet downloads are served. For example, the NuGet Gallery download URL is `https://nuget.org/api/v2/package`, so the repository URL should be configured as `https://nuget.org` and the download context path should be configured as `api/v2/package`. Default value is `api/v2/package`.
* `v3_feed_url` - (Optional) The URL to the NuGet v3 feed. Default value is `https://api.nuget.org/v3/index.json`.
* `force_nuget_authentication` - (Optional) Force basic authentication credentials in order to use this repository. Default value is `false`.
* `symbol_server_url` - (Optional) NuGet symbol server URL. Default value is `https://symbols.nuget.org/download/symbols`.
