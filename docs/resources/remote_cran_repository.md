---
subcategory: "Remote Repositories"
---
# Artifactory Remote Cran Repository Resource

Creates a remote Cran repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/CRAN+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_cran_repository" "my-remote-cran" {
  key                         = "my-remote-cran"
  url                         = "https://cran.r-project.org/"
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
$ terraform import artifactory_remote_cran_repository.my-remote-cran my-remote-cran
```
