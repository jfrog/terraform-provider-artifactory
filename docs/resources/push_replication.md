---
subcategory: "Replication"
---
# Artifactory Push Replication Resource

~> This resource is deprecated and replaced by `artifactory_local_repository_multi_replication` for clarity. 

Provides an Artifactory push replication resource. This can be used to create and manage Artifactory push replications using [Multi-push Replication API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateorReplaceLocalMulti-pushReplication).

Push replication is used to synchronize Local Repositories, and is implemented by the Artifactory server on the near
end invoking a synchronization of artifacts to the far end.

See the [Official Documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-PushReplication).

~> This resource requires Artifactory Enterprise license.

## Example Usage

```hcl
variable "artifactory_url" {
  description = "The base URL of the Artifactory deployment"
  type        = string
}

variable "artifactory_username" {
  description = "The username for the Artifactory"
  type        = string
}

variable "artifactory_password" {
  description = "The password for the Artifactory"
  type        = string
}

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
		url      = "${var.artifactory_url}/${artifactory_local_maven_repository.provider_test_dest.key}"
		username = "$var.artifactory_username"
		password = "$var.artifactory_password"
		enabled  = true
	}
}
```

## Argument Reference

The following arguments are supported:

* `repo_key` - (Required) Repository name.
* `cron_exp` - (Required) A valid CRON expression that you can use to control replication frequency. Eg: `0 0 12 * * ? *`, `0 0 2 ? * MON-SAT *`. Note: use 6 or 7 parts format - Seconds, Minutes Hours, Day Of Month, Month, Day Of Week, Year (optional). Specifying both a day-of-week AND a day-of-month parameter is not supported. One of them should be replaced by `?`. Incorrect: `* 5,7,9 14/2 * * WED,SAT *`, correct: `* 5,7,9 14/2 ? * WED,SAT *`. See details in [Cron Trigger Tutorial](https://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/crontrigger.html).
* `enable_event_replication` - (Optional) When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. added, deleted or property change.
* `replications` - (Optional)
    * `url` - (Required) The URL of the target local repository on a remote Artifactory server. Required for local repository, but not needed for remote repository.
    * `socket_timeout_millis` - (Optional) The network timeout in milliseconds to use for remote operations.
    * `username` - (Required) Required for local repository, but not needed for remote repository.
    * `password` - (Required) Required for local repository, but not needed for remote repository.
    * `enabled` - (Optional) When set, this replication will be enabled when saved.
    * `sync_deletes` - (Optional) When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata).
       Note that enabling this option, will delete artifacts on the target that do not exist in the source repository.
    * `sync_properties` - (Optional) When set, the task also synchronizes the properties of replicated artifacts.
    * `sync_statistics` - (Optional) When set, artifact download statistics will also be replicated. Set to avoid inadvertent cleanup at the target instance when setting up replication for disaster recovery.
    * `path_prefix` - (Optional) Only artifacts that located in path that matches the subpath within the remote repository will be replicated.
    * `proxy` - (Optional) Proxy key from Artifactory Proxies settings. The proxy configuration will be used when communicating with the remote instance.
    * `check_binary_existence_in_filestore` - (Optional) When true, enables distributed checksum storage. For more information, see
      [Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).

## Import

Push replication configs can be imported using their repo key, e.g.

```
$ terraform import artifactory_push_replication.foo-rep provider_test_source
```
