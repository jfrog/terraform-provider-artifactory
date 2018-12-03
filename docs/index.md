# Artifactory Provider

The [Artifactory](https://jfrog.com/artifactory/) provider is used to interact with the
resources supported by Artifactory. The provider needs to be configured
with the proper credentials before it can be used.

- Available Resources
    * [Groups](r/group.md)
    

## Example Usage

```hcl
# Configure the Artifactory provider
provider "artifactory" {
  url = "${var.artifactory_url}"
  username = "${var.artifactory_username}"
  password = "${var.artifactory_password}"
}

# Create a new repository
resource "artifactory_local_repository" "pypi-libs" {
  key             = "pypi-libs"
  package_type    = "pypi"
  repo_layout_ref = "simple-default"
  description     = "A pypi repository for python packages"
}
```

## Argument Reference

The following arguments are supported:

* `url`      - (Required) URL of Artifactory. This can also be set via the `ARTIFACTORY_URL` environment variable.
* `username` - (Optional) Username for basic auth. Requires `password` to be set. Conflicts with `token`. This can also be set via the `ARTIFACTORY_USERNAME` environment variable.
* `password` - (Optional) Password for basic auth. Requires `username` to be set. Conflicts with `token`. This can also be set via the `ARTIFACTORY_PASSWORD` environment variable.
* `username` - (Optional) API key for token auth. Uses `X-JFrog-Art-Api` header. Conflicts with `username` and `password`. This can also be set via the `ARTIFACTORY_TOKEN` environment variable.
