# Artifactory Local NPM Repository Resource

Creates a local npm repository. 

## Example Usage

```hcl
resource "artifactory_local_npm_repository" "terraform-local-test-npm-repo-basic" {
  key                 = "terraform-local-test-npm-repo-basic"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)
* `includes_pattern` - (Optional)
* `excludes_pattern` - (Optional)
* `repo_layout_ref` - (Optional)
* `checksum_policy_type` - (Optional)
* `blacked_out` - (Optional)
* `property_sets` - (Optional)

Arguments for NPM repository type closely matches with arguments for generic repository type. 