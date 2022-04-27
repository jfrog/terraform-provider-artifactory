---
subcategory: "Remote Repositories"
---
# Artifactory Remote Gitlfs Repository Resource

Creates a remote Gitlfs repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Git+LFS+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_gitlfs_repository" "my-remote-gitlfs" {
  key                         = "my-remote-gitlfs"
  url                         = "http://testartifactory.io/artifactory/example-gitlfs/"
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



## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_gitlfs_repository.my-remote-gitlfs my-remote-gitlfs
```
