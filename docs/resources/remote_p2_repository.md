---
subcategory: "Remote Repositories"
---
# Artifactory Remote P2 Repository Resource

Creates a remote P2 repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/P2+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_p2_repository" "my-remote-p2" {
  key                         = "my-remote-p2"
  url                         = "http://testartifactory.io/artifactory/example-p2/"
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
$ terraform import artifactory_remote_p2_repository.my-remote-p2 my-remote-p2
```
