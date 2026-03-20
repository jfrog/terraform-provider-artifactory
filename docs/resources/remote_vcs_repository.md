---
subcategory: "Remote Repositories"
---
# Artifactory Remote VCS Repository Resource

Creates a remote VCS repository.

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/vcs-repositories).

## Example Usage

```hcl
resource "artifactory_remote_vcs_repository" "my-remote-vcs" {
  key                  = "my-remote-vcs"
  url                  = "https://github.com/"
  vcs_git_provider     = "GITHUB"
  max_unique_snapshots = 5
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md).

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub (GITHUB), GitHub Enterprise (GITHUBENTERPRISE), BitBucket Cloud (BITBUCKET), BitBucket Server (STASH), GitLab (GITLAB), or a remote Artifactory instance. Use CUSTOM for a custom URL. Allowed values are: `GITHUB`, `GITHUBENTERPRISE`, `BITBUCKET`, `OLDSTASH`, `STASH`, `ARTIFACTORY`, `GITLAB`, `CUSTOM`. Default value is `GITHUB`.
* `vcs_git_download_url` - (Optional) This attribute is used when vcs_git_provider is set to `CUSTOM`. Provided URL will be used as proxy.
* `max_unique_snapshots` - (Optional) The maximum number of unique snapshots of a single artifact to store.
   Once the number of snapshots exceeds this setting, older versions are removed.
   A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_vcs_repository.my-remote-vcs my-remote-vcs
```
