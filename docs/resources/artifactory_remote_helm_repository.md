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
  dependencies may be downloaded from. By default, this is set to ** which means that dependencies may be downloaded
  from any external source.
