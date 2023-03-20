---
subcategory: "Replication"
---
# Artifactory Local Repository Multi Replication Resource

Provides a local repository replication resource, also referred to as Artifactory push replication. This can be used to create and manage Artifactory local repository replications using [Multi-push Replication API](https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-CreateorReplaceLocalMulti-pushReplication).
Push replication is used to synchronize Local Repositories, and is implemented by the Artifactory server on the near end invoking a synchronization of artifacts to the far end.
See the [Official Documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-PushReplication).
This resource replaces `artifactory_push_replication` and used to create a replication of one local repository to multiple repositories on the remote server. 

~> This resource requires Artifactory Enterprise license. Use `artifactory_local_repository_single_replication` with other licenses.

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

resource "artifactory_local_maven_repository" "provider_test_dest1" {
  key = "provider_test_dest1"
}

resource "artifactory_local_repository_multi_replication" "foo-rep" {
  repo_key                  = "${artifactory_local_maven_repository.provider_test_source.key}"
  cron_exp                  = "0 0 * * * ?"
  enable_event_replication  = true

	replication {
      url      = "${var.artifactory_url}/artifactory/${artifactory_local_maven_repository.provider_test_dest.key}"
      username = "$var.artifactory_username"
      password = "$var.artifactory_password"
      enabled  = true
	}
    replication {
      url      = "${var.artifactory_url}/artifactory/${artifactory_local_maven_repository.provider_test_dest1.key}"
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
* `enable_event_replication` - (Optional) When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.
* `replication` - (Optional) List of replications minimum 1 element.
    * `url` - (Required) The URL of the target local repository on a remote Artifactory server. Use the format `https://<artifactory_url>/artifactory/<repository_name>`.
    * `socket_timeout_millis` - (Optional) The network timeout in milliseconds to use for remote operations. Default value is `15000`.
    * `username` - (Required) Username on the remote Artifactory instance.
    * `password` - (Optional) Use either the HTTP authentication password or [identity token](https://www.jfrog.com/confluence/display/JFROG/User+Profile#UserProfile-IdentityTokenidentitytoken).
    * `sync_deletes` - (Optional) When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata). Note that enabling this option, will delete artifacts on the target that do not exist in the source repository. Default value is `false`.
    * `sync_properties` - (Optional) When set, the task also synchronizes the properties of replicated artifacts. Default value is `true`.
    * `sync_statistics` - (Optional) When set, the task also synchronizes artifact download statistics. Set to avoid inadvertent cleanup at the target instance when setting up replication for disaster recovery. Default value is `false`
    * `enabled` - (Optional) When set, enables replication of this repository to the target specified in `url` attribute. Default value is `true`.
    * `include_path_prefix_pattern` - (Optional) List of artifact patterns to include when evaluating artifact requests in the form of `x/y/**/z/*`. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included `(**/*)`.
    * `exclude_path_prefix_pattern` - (Optional) List of artifact patterns to exclude when evaluating artifact requests, in the form of `x/y/**/z/*`. By default, no artifacts are excluded.
    * `proxy` - (Optional) Proxy key from Artifactory Proxies settings. The proxy configuration will be used when communicating with the remote instance.
    * `replication_key` - (Computed) Replication ID, the value is unknown until the resource is created. Can't be set or updated.
    * `check_binary_existence_in_filestore` - (Optional) Enabling the `check_binary_existence_in_filestore` flag requires an Enterprise Plus license. When true, enables distributed checksum storage. For more information, see [Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).

## Import

Push replication configs can be imported using their repo key, e.g.

```
$ terraform import artifactory_local_repository_multi_replication.foo-rep provider_test_source
```
