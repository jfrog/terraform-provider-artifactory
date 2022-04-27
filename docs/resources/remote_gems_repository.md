---
subcategory: "Remote Repositories"
---
# Artifactory Remote Gems Repository Resource

Creates a remote Gems repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/RubyGems+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_gems_repository" "my-remote-gems" {
  key                         = "my-remote-gems"
  url                         = "https://rubygems.org/"
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
$ terraform import artifactory_remote_gems_repository.my-remote-gems my-remote-gems
```
