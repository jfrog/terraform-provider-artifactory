# Artifactory Remote Repository Resource

Provides an Artifactory remote `helm` repository resource. This provides helm specific fields and is the only way to get them.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Package+Management), 
although helm is (currently) not listed as a supported format

## Example Usage
Includes only new and relevant fields, for anything else, see: [generic repo](artifactory_remote_docker_repository.md).
```hcl

resource "artifactory_remote_helm_repository" "helm-remote" {
  key = "helm-remote-foo25"
  url = "https://repo.chartcenter.io/"
  helm_charts_base_url = "https://foo.com"
  external_dependencies_enabled = true
  external_dependencies_patterns = [
    "**github.com**"
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
All generic repo arguments are supported, in addition to:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `helm_charts_base_url` - (Optional) - No documentation is available. Hopefully you know what this means
* `external_dependencies_enabled` - (Optional) When set, external dependencies are rewritten.
* `external_dependencies_patterns` - (Optional) An Allow List of Ant-style path expressions that specify where external
  dependencies may be downloaded from. By default, this is an empty list which means that no dependencies may be downloaded
  from external sources. Note that the official documentation states the default is '**', which is correct when creating
  repositories in the UI, but incorrect for the API.
* `content_synchronisation` - (Optional) Reference [JFROG Smart Remote Repositories](https://www.jfrog.com/confluence/display/JFROG/Smart+Remote+Repositories)
  * `enabled` - (Optional) If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.
  * `statistics_enabled` - (Optional) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.
  * `properties_enabled` - (Optional) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.
  * `source_origin_absence_detection` - (Optional) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'
