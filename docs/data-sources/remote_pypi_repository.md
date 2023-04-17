---
subcategory: "Remote Repositories"
---
# Artifactory Remote Pypi Repository Data Source

Retrieves a remote Pypi repository.

## Example Usage

```hcl
data "artifactory_remote_pypi_repository" "remote-pypi" {
  key = "remote-pypi"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `pypi_registry_url` - (Optional) To configure the remote repo to proxy public external PyPI repository, or a PyPI repository hosted on another Artifactory server. See JFrog Pypi documentation [here](https://www.jfrog.com/confluence/display/JFROG/PyPI+Repositories) for the usage details. Default value is `https://pypi.org`.
* `pypi_repository_suffix` - (Optional) Usually should be left as a default for `simple`, unless the remote is a PyPI server that has custom registry suffix, like +simple in DevPI. Default value is `simple`.
