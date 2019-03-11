# artifactory_virtual_repository

Provides an Artifactory virtual repository resource. This can be used to create and manage Artifactory virtual repositories.

## Example Usage

```hcl
resource "artifactory_local_repository" "bar" {
  key = "bar"
  package_type = "maven"
}

resource "artifactory_local_repository" "baz" {
  key = "baz"
  package_type = "maven"
}

resource "artifactory_virtual_repository" "foo" {
  key          = "foo"
  package_type = "maven"
  repositories = [
    "${artifactory_local_repository.bar.key}", 
    "${artifactory_local_repository.baz.key}"
  ]
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Optional)
* `package_type` - (Optional)
* `repositories` - (Optional)
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional)
* `excludes_pattern` - (Optional)
* `repo_layout_ref` - (Optional)
* `debian_trivial_layout` - (Optional)
* `artifactory_requests_can_retrieve_remote_artifacts` - (Optional)
* `key_pair` - (Optional)
* `pom_repository_references_cleanup_policy` - (Optional)
* `default_deployment_repo` - (Optional)

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_repository.foo foo
```