# Artifactory Remote Opkg Repository Resource

Creates a remote Opkg repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Opkg+Repositories)


## Example Usage
To create a new Artifactory remote Opkg repository called my-remote-opkg.

```hcl
resource "artifactory_remote_opkg_repository" "my-remote-opkg" {
  key                         = "my-remote-opkg"
  url                         = "http://testartifactory.io/artifactory/example-opkg/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Opkg repository type closely match with arguments for remote Generic repository type.