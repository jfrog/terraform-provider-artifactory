# Artifactory Remote Go Repository Resource

Creates a remote VCS repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/VCS+Repositories)


## Example Usage
To create a new Artifactory remote VCS repository called my-remote-vcs.

```hcl
resource "artifactory_remote_go_repository" "my-remote-vcs" {
  key                         = "my-remote-vcs"
  url                         = "https://github.com/"
  vcs_git_provider            = "GITHUB"
  max_unique_snapshots        = 5
}
```

## Argument Reference

Arguments have a one to one mapping with the
[JFrog API](https://www.jfrog.com/confluence/display/JFROG/Repository+Configuration+JSON#RepositoryConfigurationJSON-RemoteRepository).
The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL.
* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub, Bitbucket, 
   Stash, a remote Artifactory instance or a custom Git repository. Allowed values are: 'GITHUB', 'BITBUCKET', 'OLDSTASH', 
   'STASH', 'ARTIFACTORY', 'CUSTOM'. Default value is 'GITHUB'
* `vcs_git_download_url` - (Optional) This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.
* `max_unique_snapshots` - (Optional) - The maximum number of unique snapshots of a single artifact to store.
   Once the number of snapshots exceeds this setting, older versions are removed.
   A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.

Arguments for remote VCS repository type closely match with arguments for remote Generic repository type.