---
subcategory: "Remote Repositories"
---
# Artifactory Remote Bower Repository Resource

Creates a remote Bower repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Bower+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_bower_repository" "my-remote-bower" {
  key                         = "my-remote-bower"
  url                         = "https://github.com/"
  vcs_git_provider            = "GITHUB"
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
* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is `GITHUB`. Possible values are: `GITHUB`, `BITBUCKET`, `OLDSTASH`, `STASH`, `ARTIFACTORY`, `CUSTOM`.
* `vcs_git_download_url` - (Optional) This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.
* `bower_registry_url` - (Optional) Proxy remote Bower repository. Default value is `https://registry.bower.io`.



## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_bower_repository.my-remote-bower my-remote-bower
```
