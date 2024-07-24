---
subcategory: "Remote Repositories"
---
# Artifactory Remote Ansible Repository Resource

Creates a remote Ansible repository.

Official documentation can be found [here](https://jfrog.com/help/r/jfrog-artifactory-documentation/ansible-repositories).


## Example Usage

```terraform
resource "artifactory_remote_alpine_repository" "my-remote-ansible" {
  key = "my-remote-ansible"
  url = "https://galaxy.ansible.com"
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

## Import

Remote repositories can be imported using their name, e.g.
```shell
terraform import artifactory_remote_ansible_repository.my-remote-ansible my-remote-ansible
```
