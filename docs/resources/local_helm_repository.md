---
subcategory: "Local Repositories"
---
# Artifactory Local Helm Repository Resource

Creates a local Helm repository.

## Example Usage

```hcl
resource "artifactory_local_helm_repository" "terraform-local-test-helm-repo" {
  key = "terraform-local-test-helm-repo"
  force_non_duplicate_chart = true
  force_metadata_name_version = false
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the local repositories](local.md):

* `key` - (Required) the identity key of the repo.
* `description` - (Optional)
* `notes` - (Optional)
* `force_non_duplicate_chart` - (Optional) Prevents the deployment of charts with the same name and version in different repository paths. Only available for 7.104.0 onward. Cannot be updated after it is set.
* `force_metadata_name_version` - (Optional) Ensures that the chart name and version in the file name match the values in Chart.yaml and adhere to SemVer standards. Only available for 7.104.0 onward. Cannot be updated after it is set.

## Import

Local repositories can be imported using their name, e.g.
```
$ terraform import artifactory_local_helm_repository.terraform-local-test-helm-repo terraform-local-test-helm-repo
```
