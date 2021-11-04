# Artifactory Local Docker V1 Repository Resource
Creates a local docker v1 repository - By choosing a V1 repository, you don't really have many options 

## Example Usage

```hcl
resource "artifactory_local_docker_v2_repository" "foo" {
  key 	     = "foo"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON). The following arguments are supported:

* `key` - (Required) - the identity key of the repo
