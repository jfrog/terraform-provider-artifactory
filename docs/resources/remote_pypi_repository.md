---
subcategory: "Remote Repositories"
---
# Artifactory Remote Repository Resource

Creates a remote Pypi repository.
Official documentation can be found [here](https://www.jfrog.com/confluence/display/JFROG/PyPI+Repositories).

## Example Usage

```hcl
resource "artifactory_remote_pypi_repository" "pypi-remote" {
  key                    = "pypi-remote-foo"
  url                    = "https://files.pythonhosted.org"
  pypi_registry_url      = "https://pypi.org"
  pypi_repository_suffix = "simple"
}
```

## Argument Reference

Arguments have a one to one mapping with the [JFrog API](https://www.jfrog.com/confluence/display/RTF/Repository+Configuration+JSON).
The following arguments are supported, along with the [common list of arguments for the remote repositories](remote.md):

* `key` - (Required) A mandatory identifier for the repository that must be unique. It cannot begin with a number or
  contain spaces or special characters.
* `description` - (Optional)
* `notes` - (Optional)
* `url` - (Required) The remote repo URL.
* `pypi_registry_url` - (Optional) To configure the remote repo to proxy public external PyPI repository, or a PyPI repository hosted on another Artifactory server. See JFrog Pypi documentation [here](https://www.jfrog.com/confluence/display/JFROG/PyPI+Repositories) for the usage details. Default value is `https://pypi.org`.
* `pypi_repository_suffix` - (Optional) Usually should be left as a default for `simple`, unless the remote is a PyPI server that has custom registry suffix, like +simple in DevPI. Default value is `simple`.

## Import

Remote repositories can be imported using their name, e.g.
```
$ terraform import artifactory_remote_pypi_repository.pypi-remote pypi-remote
```
