---
subcategory: "Remote Repositories"
---
# Artifactory Remote Go Repository Resource

Creates a remote Go repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Go+Registry).


## Example Usage

```hcl
resource "artifactory_remote_go_repository" "my-remote-go" {
  key                         = "my-remote-go"
  url                         = "https://proxy.golang.org/"
  vcs_git_provider            = "ARTIFACTORY"
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
* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is `ARTIFACTORY`.



## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_go_repository.my-remote-go my-remote-go
```
