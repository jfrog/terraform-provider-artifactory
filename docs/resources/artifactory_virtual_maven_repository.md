# Artifactory Virtual Maven Repository Resource

Provides an Artifactory virtual repository resource with specific maven feature.

## Example Usage

```hcl
resource "artifactory_local_maven_repository" "bar" {
  key = "bar"
  repo_layout_ref = "maven-2-default"
}

resource "artifactory_remote_maven_repository" "baz" {
  key             = "baz"
  url             = "https://search.maven.com/"
  repo_layout_ref = "maven-2-default"
}

resource "artifactory_virtual_maven_repository" "foo" {
  key          = "maven-virt-repo"
  repo_layout_ref = "maven-2-default"
  repositories = [
    "${artifactory_local_maven_repository.bar.key}",
    "${artifactory_local_maven_repository.baz.key}"
  ]
  description = "A test virtual repo"
  notes = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
  force_maven_authentication = true
  pom_repository_references_cleanup_policy = "discard_active_reference"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required)
* `pom_repository_references_cleanup_policy` - (Optional). One of: `"discard_active_reference", "discard_any_reference", "nothing"`
* `force_maven_authentication` - (Optional) - forces authentication when fetching from remote repos

Arguments for Maven repository type closely match with arguments for Generic repository type.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_maven_repository.foo foo
```
