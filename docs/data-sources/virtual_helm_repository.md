---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Helm Repository Data Source

Retrieves a virtual Helm repository.

## Example Usage

```hcl
data "artifactory_virtual_helm_repository" "virtual-helm" {
  key = "virtual-helm"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of arguments for the virtual repositories](../resources/virtual.md):

* `retrieval_cache_period_seconds` - (Optional, Default: `7200`) This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.
* `use_namespaces` - (Optional) From Artifactory 7.24.1 (SaaS Version), you can explicitly state a specific aggregated local or remote repository to fetch from a virtual by assigning namespaces to local and remote repositories. See the documentation [here](https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories#KubernetesHelmChartRepositories-NamespaceSupportforHelmVirtualRepositories). Default is `false`.
