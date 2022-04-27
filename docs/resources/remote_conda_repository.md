---
subcategory: "Remote Repositories"
---
# Artifactory Remote Conda Repository Resource

Creates a remote Conda repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Conda+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_conda_repository" "my-remote-conda" {
  key                         = "my-remote-conda"
  url                         = "https://repo.anaconda.com/pkgs/main"
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
$ terraform import artifactory_remote_conda_repository.my-remote-conda my-remote-conda
```
