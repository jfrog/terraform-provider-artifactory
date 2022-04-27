---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual Ivy Repository Resource

Creates a virtual Ivy repository.
Official documentation can be found [here](https://jfrog.com/blog/how-to-set-up-a-private-remote-and-virtual-maven-gradle-registry/).

## Example Usage

```hcl
resource "artifactory_virtual_ivy_repository" "foo-ivy" {
  key                                       = "foo-ivy"
  repositories                              = []
  description                               = "A test virtual repo"
  notes                                     = "Internal description"
  includes_pattern                          = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                          = "com/google/**"
  pom_repository_references_cleanup_policy  = "discard_active_reference"
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
* `pom_repository_references_cleanup_policy` - (Optional)
    - (1: discard_active_reference) Discard Active References - Removes repository elements that are declared directly under project or under a profile in the same POM that is activeByDefault.
    - (2: discard_any_reference) Discard Any References - Removes all repository elements regardless of whether they are included in an active profile or not.
    - (3: nothing) Nothing - Does not remove any repository elements declared in the POM.
* `key_pair` - (Optional) The keypair used to sign artifacts.


## Import

Virtual repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_ivy_repository.foo-ivy foo-ivy
```
