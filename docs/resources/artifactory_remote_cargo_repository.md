# Artifactory Remote Cargo Repository Resource

Provides an Artifactory remote `cargo` repository resource. This provides cargo specific fields and is the only way to get them
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Cargo+Registry)


## Example Usage
Create a new Artifactory remote cargo repository called my-remote-cargo
for brevity sake, only cargo specific fields are included; for other fields see documentation for
[generic repo](artifactory_remote_docker_repository.md).
```hcl

resource "artifactory_remote_cargo_repository" "my-remote-cargo" {
  key                 = "my-remote-cargo"
  anonymous_access    = true
  git_registry_url    = "https://github.com/rust-lang/foo.index"
}
```
## Note
If you get a 400 error: `"Custom Base URL should be defined prior to creating a Cargo repository"`,
you must set the base url at: `http://${host}/ui/admin/configuration/general`

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `project_key` - (Optional) Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: "DEV" or "PROD"
* `anonymous_access` - (Required) - Cargo client does not send credentials when performing download and search for crates. Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option.
* `git_registry_url` - (Optional) - This is the index url, expected to be a git repository. for remote artifactory use "arturl/git/repokey.git"
* `content_synchronisation` - (Optional) Reference [JFROG Smart Remote Repositories](https://www.jfrog.com/confluence/display/JFROG/Smart+Remote+Repositories)
    * `enabled` - (Optional) If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.
    * `statistics_enabled` - (Optional) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.
    * `properties_enabled` - (Optional) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.
    * `source_origin_absence_detection` - (Optional) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'
* `xray_index` - (Optional, Default: false)  Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.
