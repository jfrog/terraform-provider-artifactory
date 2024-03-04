---
subcategory: "Virtual Repositories"
---
# Artifactory Virtual OCI Repository Resource

Creates a virtual OCI repository.

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/set-up-virtual-oci-repositories).

## Example Usage

```hcl
resource "artifactory_virtual_oci_repository" "my-oci-virtual" {
  key                            = "my-oci-virtual"
  repositories                   = ["my-oci-local", "my-oci-remote"]
  description                    = "A test virtual OCI repo"
  notes                          = "Internal description"
  resolve_oci_tags_by_timestamp  = true
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). 

The following arguments are supported, along with the [common list of arguments for the virtual repositories](virtual.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.
* `repositories` - (Optional) The effective list of actual repositories included in this virtual repository.
* `description` - (Optional)
* `notes` - (Optional)
* `resolve_oci_tags_by_timestamp` - (Optional) When enabled, in cases where the same OCI tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp. Default values is `false`.

## Import

Virtual OCI repositories can be imported using their name, e.g.

```
$ terraform import artifactory_virtual_oci_repository.my-oci-virtual my-oci-virtual
```
