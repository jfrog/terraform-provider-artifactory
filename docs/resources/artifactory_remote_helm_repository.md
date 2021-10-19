# Artifactory Remote Repository Resource

Provides an Artifactory remote `helm` repository resource. This provides helm specific fields and is the only way to get them.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Package+Management), 
although helm is (currently) not listed as a supported format


## Example Usage
## Example Usage
Includes only new and relevant fields, for anything else, see: [generic repo](artifactory_remote_docker_repository.md).
```hcl

resource "artifactory_remote_helm_repository" "helm-remote" {
  key = "helm-remote-foo25"
  url = "https://repo.chartcenter.io/"
  helm_charts_base_url = "https://foo.com"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
All generic repo arguments are supported, in addition to:

* `key` - (Required) The repository identifier. Must be unique system-wide
* `helm_charts_base_url` - (Optional) - No documentation is available. Hopefully you know what this means
