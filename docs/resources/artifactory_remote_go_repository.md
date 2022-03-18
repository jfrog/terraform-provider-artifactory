# Artifactory Remote Go Repository Resource

Creates a remote Go repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Go+Registry)


## Example Usage
To create a new Artifactory remote Go repository called my-remote-go.

```hcl
resource "artifactory_remote_go_repository" "my-remote-go" {
  key                         = "my-remote-go"
  url                         = "https://proxy.golang.org/"
  vcs_git_provider            = "ARTIFACTORY"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) - the remote repo URL. You kinda don't have a remote repo without it
* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "ARTIFACTORY".

Arguments for remote Go repository type closely match with arguments for remote Generic repository type.