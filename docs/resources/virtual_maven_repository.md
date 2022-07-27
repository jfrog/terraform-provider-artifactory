---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Maven Repository Resource

Creates a virtual Maven repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/Maven+Repository).

## Example Usage

```hcl
resource "artifactory_local_maven_repository" "bar" {
  key             = "bar"
  repo_layout_ref = "maven-2-default"
}

resource "artifactory_remote_maven_repository" "baz" {
  key             = "baz"
  url             = "https://search.maven.com/"
  repo_layout_ref = "maven-2-default"
}

resource "artifactory_virtual_maven_repository" "maven-virt-repo" {
  key             = "maven-virt-repo"
  repo_layout_ref = "maven-2-default"
  repositories    = [
    "${artifactory_local_maven_repository.bar.key}",
    "${artifactory_remote_maven_repository.baz.key}"
  ]
  description                              = "A test virtual repo"
  notes                                    = "Internal description"
  includes_pattern                         = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                         = "com/google/**"
  force_maven_authentication               = true
  pom_repository_references_cleanup_policy = "discard_active_reference"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `pom_repository_references_cleanup_policy` - (Optional) One of: `"discard_active_reference", "discard_any_reference", "nothing"`
* `force_maven_authentication` - (Optional) Forces authentication when fetching from remote repos.

## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_maven_repository.maven-virt-repo maven-virt-repo
```
