---
subcategory: "Replication"
---
# Artifactory Single Replication Config Resource

Note: this resource is deprecated in favor of `artifactory_pull_replication` resource.

Provides an Artifactory single replication config resource. This can be used to create and manage a single Artifactory
replication. Primarily used when pull replication is needed.

**WARNING: This should not be used on a repository with `artifactory_replication_config`. Using both together will cause
unexpected behaviour and will almost certainly cause your replications to break.**

### Passwords
Passwords can only be used when encryption is turned off, see [Key Encryption documentation](https://www.jfrog.com/confluence/display/RTF/Artifactory+Key+Encryption). 
Since only the artifactory server can decrypt them it is impossible for terraform to diff changes correctly.

To get full management, passwords can be decrypted globally using `POST /api/system/decrypt`. If this is not possible, 
the password diff can be disabled per resource with-- noting that this will require resources to be tainted for an update:
```hcl
lifecycle {
    ignore_changes = ["password"]
}
``` 

## Example Usage

```hcl
# Create a replication between two artifactory local repositories
resource "artifactory_local_maven_repository" "provider_test_source" {
	key = "provider_test_source"
}

resource "artifactory_local_maven_repository" "provider_test_dest" {
	key = "provider_test_dest"
}

resource "artifactory_single_replication_config" "foo-rep" {
	repo_key                  = "${artifactory_local_maven_repository.provider_test_source.key}"
	cron_exp                  = "0 0 * * * ?"
	enable_event_replication  = true
    url                       = "${var.artifactory_url}"
    username                  = "${var.artifactory_username}"
    password                  = "${var.artifactory_password}"		
}
```

## Argument Reference

The following arguments are supported:

* `repo_key` - (Required)
* `cron_exp` - (Required)
* `enable_event_replication` - (Optional)
* `url` - (Required)
* `socket_timeout_millis` - (Optional)
* `username` - (Optional)
* `password` - (Optional) Requires password encryption to be turned off `POST /api/system/decrypt`.
* `enabled` - (Optional)
* `sync_deletes` - (Optional)
* `sync_properties` - (Optional)
* `sync_statistics` - (Optional)
* `path_prefix` - (Optional)
* `proxy` - (Optional) Proxy key from Artifactory Proxies setting.

## Import

Replication configs can be imported using their repo key, e.g.

```
$ terraform import artifactory_single_replication_config.foo-rep repository-key
```
