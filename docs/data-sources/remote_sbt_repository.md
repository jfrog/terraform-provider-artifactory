---
subcategory: "Remote Repositories"
---
# Artifactory Remote SBT Repository Data Source

Retrieves a remote SBT repository.

## Example Usage

```hcl
data "artifactory_remote_sbt_repository" "remote-sbt" {
  key = "remote-sbt"
}
```

## Argument Reference

The following argument is supported:

* `key` - (Required) the identity key of the repo.

## Attribute Reference

The following attributes are supported, along with the [common list of attributes for the remote repositories](../resources/remote.md):

* `fetch_jars_eagerly` - (Optional, Default: `false`) When set, if a POM is requested, Artifactory attempts to fetch the corresponding jar in the background. This will accelerate first access time to the jar when it is subsequently requested. 
* `fetch_sources_eagerly` - (Optional, Default: `false`) - When set, if a binaries jar is requested, Artifactory attempts to fetch the corresponding source jar in the background. This will accelerate first access time to the source jar when it is subsequently requested.
* `handle_releases` - (Optional, Default: `true`) If set, Artifactory allows you to deploy release artifacts into this repository.
* `handle_snapshots` - (Optional, Default: `true`) If set, Artifactory allows you to deploy snapshot artifacts into this repository.
* `suppress_pom_consistency_checks` - (Optional, Default: `true`) By default, the system keeps your repositories healthy by refusing POMs with incorrect coordinates (path). If the groupId:artifactId:version information inside the POM does not match the deployed path, Artifactory rejects the deployment with a "409 Conflict" error. You can disable this behavior by setting this attribute to `true`.
* `reject_invalid_jars` - (Optional, Default: `false`) Reject the caching of jar files that are found to be invalid. For example, pseudo jars retrieved behind a "captive portal".
* `remote_repo_checksum_policy_type` - (Optional, Default: `generate-if-absent`) Checking the Checksum effectively verifies the integrity of a deployed resource. The Checksum Policy determines how the system behaves when a client checksum for a remote resource is missing or conflicts with the locally calculated checksum. Available policies are `generate-if-absent`, `fail`, `ignore-and-generate`, and `pass-thru`. 