# Artifactory Push Replication Resource

Provides an Artifactory push replication resource. This can be used to create and manage Artifactory push replications.


## Example Usage

```hcl
# Create a replication between two artifactory local repositories
resource "artifactory_local_maven_repository" "provider_test_source" {
	key = "provider_test_source"
}

resource "artifactory_local_maven_repository" "provider_test_dest" {
	key = "provider_test_dest"
}

resource "artifactory_push_replication" "foo-rep" {
	repo_key                  = "${artifactory_local_maven_repository.provider_test_source.key}"
	cron_exp                  = "0 0 * * * ?"
	enable_event_replication  = true

	replications {
		url      = "$var.artifactory_url"
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
    * `username` - (Required)
    * `password` - (Required)
    * `enabled` - (Optional)
    * `sync_deletes` - (Optional)
    * `sync_properties` - (Optional)
    * `sync_statistics` - (Optional)
    * `path_prefix` - (Optional)
    * `proxy` - (Optional) Proxy key from Artifactory Proxies setting.

## Import

Push replication configs can be imported using their repo key, e.g.

```
$ terraform import artifactory_push_replication.foo-rep provider_test_source
```
