# Artifactory Remote Alpine Repository Resource

Creates a remote Alpine repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories)


## Example Usage
To create a new Artifactory remote Alpine repository called my-remote-alpine.

```hcl
resource "artifactory_remote_alpine_repository" "my-remote-alpine" {
  key                         = "my-remote-alpine"
  url                         = "http://dl-cdn.alpinelinux.org/alpine"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Alpine repository type closely match with arguments for remote Generic repository type.