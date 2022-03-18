# Artifactory Remote Nuget Repository Resource

Creates a remote Nuget repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/NuGet+Repositories)


## Example Usage
To create a new Artifactory remote Nuget repository called my-remote-nuget.

```hcl
resource "artifactory_remote_nuget_repository" "my-remote-nuget" {
  key                         = "my-remote-nuget"
  url                         = "https://www.nuget.org/"
  download_context_path       = "api/v2/package"
  force_nuget_authentication  = true
  v3_feed_url                 = "https://api.nuget.org/v3/index.json"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it
* `feed_context_path` - (Optional) When proxying a remote NuGet repository, customize feed resource location using this attribute. Default value is 'api/v2'.
* `download_context_path` - (Optional) The context path prefix through which NuGet downloads are served. Default value is 'api/v2/package'.
* `v3_feed_url` - (Optional) The URL to the NuGet v3 feed. Default value is 'https://api.nuget.org/v3/index.json'.
* `force_nuget_authentication` - (Optional) Force basic authentication credentials in order to use this repository. Default value is 'false'.

Arguments for remote Nuget repository type closely match with arguments for remote Generic repository type.