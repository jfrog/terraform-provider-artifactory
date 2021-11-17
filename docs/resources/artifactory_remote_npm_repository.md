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
* `list_remote_folder_items` - (Optional) - No documentation could be found. This field exist in the API but not in the UI
* `mismatching_mime_types_override_list` - (Optional) - No documentation could be found. This field exist in the API but not in the UI


