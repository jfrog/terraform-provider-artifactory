---
subcategory: "Remote Repositories"
---
# Artifactory Remote VCS Repository Data Source

Retrieves a remote VCS repository.

## Example Usage

```hcl
data "artifactory_remote_vcs_repository" "remote-vcs" {
  key = "remote-vcs"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub (GITHUB), GitHub Enterprise (GITHUBENTERPRISE), BitBucket Cloud (BITBUCKET), BitBucket Server (STASH), GitLab (GITLAB), or a remote Artifactory instance. Use CUSTOM for a custom URL. Allowed values are: `GITHUB`, `GITHUBENTERPRISE`, `BITBUCKET`, `OLDSTASH`, `STASH`, `ARTIFACTORY`, `GITLAB`, `CUSTOM`. Default value is `GITHUB`.
* `vcs_git_download_url` - (Optional) This attribute is used when vcs_git_provider is set to `CUSTOM`. Provided URL will be used as proxy.
* `max_unique_snapshots` - (Optional) The maximum number of unique snapshots of a single artifact to store. Once the number of snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.
