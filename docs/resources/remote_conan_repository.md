---
subcategory: "Remote Repositories"
---
# Artifactory Remote Conan Repository Resource

Creates a remote Conan repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Conan+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_conan_repository" "my-remote-conan" {
  key                         = "my-remote-conan"
  url                         = "https://conan.io/center/"
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
$ terraform import artifactory_remote_conan_repository.my-remote-conan my-remote-conan
```
