# Artifactory Local Nuget Repository Resource

Creates a local Nuget repository and allows for the creation of a 

## Example Usage

```hcl
resource "artifactory_local_nuget_repository" "terraform-local-test-nuget-repo-basic" {
  key                 = "terraform-local-test-nuget-repo-basic"
  max_unique_snapshots = 5
  force_nuget_authentication = true
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `max_unique_snapshots` - (Optional) - The maximum number of unique snapshots of a single artifact to store.
  Once the number of snapshots exceeds this setting, older versions are removed.
  A value of 0 (default) indicates there is no limit, and unique snapshots are not cleaned up.
* `force_nuget_authentication` - (Optional) - Force basic authentication credentials in order to use this repository.