# Artifactory Remote Gitlfs Repository Resource

Creates a remote Gitlfs repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Git+LFS+Repositories)


## Example Usage
To create a new Artifactory remote Gitlfs repository called my-remote-gitlfs.

```hcl
resource "artifactory_remote_gitlfs_repository" "my-remote-gitlfs" {
  key                         = "my-remote-gitlfs"
  url                         = "http://testartifactory.io/artifactory/example-gitlfs/"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it

Arguments for remote Gitlfs repository type closely match with arguments for remote Generic repository type.