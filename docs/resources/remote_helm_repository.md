---
subcategory: "Remote Repositories"
---
# Artifactory Remote Repository Resource

Provides a remote Helm repository. 
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories).


## Example Usage

```hcl
resource "artifactory_remote_helm_repository" "helm-remote" {
  key                             = "helm-remote-foo25"
  url                             = "https://repo.chartcenter.io/"
  helm_charts_base_url            = "https://foo.com"
  external_dependencies_enabled   = true
  external_dependencies_patterns  = [
    "**github.com**"
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `helm_charts_base_url` - (Optional) Base URL for the translation of chart source URLs in the index.yaml of virtual repos. Artifactory will only translate URLs matching the index.yamls hostname or URLs starting with this base url. Support http/https/oci protocol scheme.
* `external_dependencies_enabled` - (Optional) When set, external dependencies are rewritten. `External Dependency Rewrite` in the UI.
* `external_dependencies_patterns` - (Optional) An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will
  follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response.
  By default, this is set to `[**]` in the UI, which means that remote modules may be downloaded from any external VCS source.
  Due to SDKv2 limitations, we can't set the default value for the list.
  This value `[**]` must be assigned to the attribute manually, if user don't specify any other non-default values.
  We don't want to make this attribute required, but it must be set to avoid the state drift on update. Note: Artifactory assigns
  `[**]` on update if HCL doesn't have the attribute set or the list is empty.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_helm_repository.helm-remote helm-remote
```
