# Artifactory Remote Debian Repository Resource

Creates a remote Debian repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Debian+Repositories)


## Example Usage
To create a new Artifactory remote Debian repository called my-remote-debian.

```hcl
resource "artifactory_remote_debian_repository" "my-remote-debian" {
  key                         = "my-remote-Debian"
  url                         = "http://archive.ubuntu.com/ubuntu/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Debian repository type closely match with arguments for remote Generic repository type.