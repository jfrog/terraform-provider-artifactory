---
layout: "artifactory"
page_title: "Artifactory: artifactory_replication_config"
sidebar_current: "docs-artifactory-resource-replication-config"
description: |-
  Provides an replication config resource.
---

# artifactory_replication_config

Provides an Artifactory replication config resource. This can be used to create and manage Artifactory replications.

### Passwords
Passwords can only be used when encryption is turned off (https://www.jfrog.com/confluence/display/RTF/Artifactory+Key+Encryption). 
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
resource "artifactory_local_repository" "provider_test_source" {
	key = "provider_test_source"
	package_type = "maven"
}

resource "artifactory_local_repository" "provider_test_dest" {
	key = "provider_test_dest"
	package_type = "maven"
}

resource "artifactory_replication_config" "foo-rep" {
	repo_key = "${artifactory_local_repository.provider_test_source.key}"
	cron_exp = "0 0 * * * ?"
	enable_event_replication = true
	
	replications {
		url = "$var.artifactory_url"
		username = "$var.artifactory_username"
		password = "$var.artifactory_password"
	}
}
```

## Argument Reference

The following arguments are supported:

* `repo_key` - (Required)
* `cron_exp` - (Required)
* `enable_event_replication` - (Optional)
* `replications` - (Optional)
    * `url` - (Required)
    * `socket_timeout_millis` - (Optional)
    * `username` - (Optional)
    * `password` - (Optional) Requires password encryption to be turned off `POST /api/system/decrypt`
    * `enabled` - (Optional)
    * `sync_deletes` - (Optional)
    * `sync_properties` - (Optional)
    * `sync_statistics` - (Optional)
    * `path_prefix` - (Optional)

## Import

Replication configs can be imported using their repo key, e.g.

```
$ terraform import artifactory_replication_config.foo-rep provider_test_source
```