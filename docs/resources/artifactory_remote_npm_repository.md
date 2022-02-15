# Artifactory Remote npm Repository Resource

Provides an Artifactory remote `npm` repository resource. This provides npm specific fields and is the only way to get them
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/npm+Registry)


## Example Usage
Create a new Artifactory remote npm repository called my-remote-npm
for brevity sake, only npm specific fields are included; for other fields see documentation for
[generic repo](artifactory_remote_docker_repository.md).
```hcl

resource "artifactory_remote_npm_repository" "thing" {
  key                         = "remote-thing-npm"
  url                         = "https://registry.npmjs.org/"
  list_remote_folder_items    = true
  mismatching_mime_types_override_list = "application/json,application/xml"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `project_key` - (Optional) Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.
* `project_environments` - (Optional) Project environment for assigning this repository to. Allow values: "DEV" or "PROD"
* `list_remote_folder_items` - (Optional) - No documentation could be found. This field exist in the API but not in the UI
* `mismatching_mime_types_override_list` - (Optional) - No documentation could be found. This field exist in the API but not in the UI
* `content_synchronisation` - (Optional) Reference [JFROG Smart Remote Repositories](https://www.jfrog.com/confluence/display/JFROG/Smart+Remote+Repositories)
    * `enabled` - (Optional) If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.
    * `statistics_enabled` - (Optional) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.
    * `properties_enabled` - (Optional) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.
    * `source_origin_absence_detection` - (Optional) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'
