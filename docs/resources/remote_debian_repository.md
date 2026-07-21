---
subcategory: "Remote Repositories"
---
# Artifactory Remote Debian Repository Resource

Creates a remote Debian repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Debian+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_debian_repository" "my-remote-debian" {
  key          = "my-remote-debian"
  url          = "http://archive.ubuntu.com/ubuntu/"
  curated      = true
  pass_through = false
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
* `curated` - (Optional, Default: `false`) Enable repository to be protected by the Curation service.
* `pass_through` - (Optional, Default: `false`) Enable Pass-through for Curation Audit. When enabled, allows artifacts to pass through the Curation audit process.



## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_debian_repository.my-remote-debian my-remote-debian
```
