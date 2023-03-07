---
subcategory: "Remote Repositories"
---
# Artifactory Remote Nuget Repository Resource

Creates a remote Nuget repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/NuGet+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_nuget_repository" "my-remote-nuget" {
  key                         = "my-remote-nuget"
  url                         = "https://www.nuget.org/"
  download_context_path       = "api/v2/package"
  force_nuget_authentication  = true
  v3_feed_url                 = "https://api.nuget.org/v3/index.json"
  symbol_server_url           = "https://symbols.nuget.org/download/symbols"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `feed_context_path` - (Optional) When proxying a remote NuGet repository, customize feed resource location using this attribute. Default value is `api/v2`.
* `download_context_path` - (Optional) The context path prefix through which NuGet downloads are served.
   For example, the NuGet Gallery download URL is `https://nuget.org/api/v2/package`, so the repository
   URL should be configured as `https://nuget.org` and the download context path should be configured as `api/v2/package`. Default value is `api/v2/package`.
* `v3_feed_url` - (Optional) The URL to the NuGet v3 feed. Default value is `https://api.nuget.org/v3/index.json`.
* `force_nuget_authentication` - (Optional) Force basic authentication credentials in order to use this repository. Default value is `false`.
* `symbol_server_url` - (Optional) NuGet symbol server URL. Default value is `https://symbols.nuget.org/download/symbols`.



## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_nuget_repository.my-remote-nuget my-remote-nuget
```
