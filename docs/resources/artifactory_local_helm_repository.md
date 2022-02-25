# Artifactory Local Helm Repository Resource

Creates a local helm repository.

## Example Usage

```hcl
resource "artifactory_local_helm_repository" "terraform-local-test-helm-repo" {
  key = "terraform-local-test-helm-repo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
* `description` - (Optional)
* `notes` - (Optional)
