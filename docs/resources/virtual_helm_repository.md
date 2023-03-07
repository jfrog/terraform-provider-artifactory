---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Helm Repository Resource

Creates a virtual Helm repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories#KubernetesHelmChartRepositories-VirtualRepositories).

## Example Usage

```hcl
resource "artifactory_virtual_helm_repository" "foo-helm-virtual" {
  key            = "foo-helm-virtual"
  use_namespaces = true
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 
The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `notes` - (Optional)
* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.
* `use_namespaces` - (Optional) From Artifactory 7.24.1 (SaaS Version), you can explicitly state a specific aggregated local or remote repository to fetch from a virtual by assigning namespaces to local and remote repositories. See the documentation [here](https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories#KubernetesHelmChartRepositories-NamespaceSupportforHelmVirtualRepositories). Default is `false`.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_helm_repository.foo-helm-virtual foo-helm-virtual
```
