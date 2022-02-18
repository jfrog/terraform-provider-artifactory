# Artifactory Virtual Helm Repository Resource

Provides an Artifactory virtual repository resource with Helm package type. This should be preferred over the original one-size-fits-all `artifactory_virtual_repository`.

## Example Usage

```hcl
resource "artifactory_virtual_helm_repository" "foo-helm-virtual" {
  key            = "foo-helm-virtual"
  use_namespaces = true
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/JFROG/Repository+Configuration+JSON#RepositoryConfigurationJSON-VirtualRepository). The following arguments are supported:

* `key` - (Required)
* `use_namespaces` - (Optional) - From Artifactory 7.24.1 (SaaS Version), you can explicitly state a specific aggregated local or remote repository to fetch from a virtual by assigning namespaces to local and remote repositories. See https://www.jfrog.com/confluence/display/JFROG/Kubernetes+Helm+Chart+Repositories#KubernetesHelmChartRepositories-NamespaceSupportforHelmVirtualRepositories. Default to 'false'.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_helm_repository.foo foo
```
