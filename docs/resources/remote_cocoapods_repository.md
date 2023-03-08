---
subcategory: "Remote Repositories"
---
# Artifactory Remote CocoaPods Repository Resource

Creates a remote CocoaPods repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/CocoaPods+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_cocoapods_repository" "my-remote-cocoapods" {
  key                         = "my-remote-cocoapods"
  url                         = "https://github.com/"
  vcs_git_provider            = "GITHUB"
  pods_specs_repo_url         = "https://github.com/CocoaPods/Spec"
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
* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is `GITHUB`. 
   Possible values are: `GITHUB`, `BITBUCKET`, `OLDSTASH`, `STASH`, `ARTIFACTORY`, `CUSTOM`.
* `vcs_git_download_url` - (Optional) This attribute is used when vcs_git_provider is set to `CUSTOM`. Provided URL will be used as proxy.
* `pods_specs_repo_url` - (Optional) Proxy remote CocoaPods Specs repositories. Default value is `https://github.com/CocoaPods/Specs`.



## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_cocoapods_repository.my-remote-cocoapods my-remote-cocoapods
```
