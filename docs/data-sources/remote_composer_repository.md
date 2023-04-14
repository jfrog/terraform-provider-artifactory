---
subcategory: "Remote Repositories"
---
# Artifactory Remote Composer Repository Data Source

Retrieves a remote Composer repository.

## Example Usage

```hcl
data "artifactory_remote_composer_repository" "remote-composer" {
  key = "remote-composer"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `vcs_git_provider` - (Optional) Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is `GITHUB`. Possible values are: `GITHUB`, `BITBUCKET`, `OLDSTASH`, `STASH`, `ARTIFACTORY`, `CUSTOM`.
* `vcs_git_download_url` - (Optional) This attribute is used when vcs_git_provider is set to `CUSTOM`. Provided URL will be used as proxy.
* `composer_registry_url` - (Optional) Proxy remote Composer repository. Default value is `https://packagist.org`.
