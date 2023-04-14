## 7.6.0 (April 14, 2023). Tested on Artifactory 7.55.10

FEATURES:

* datasource/artifactory_virtual_*_repository: Adds new data sources for all virtual repository package types.
  PR:     [#719](https://github.com/jfrog/terraform-provider-artifactory/pull/719)
  Issues: [#548](https://github.com/jfrog/terraform-provider-artifactory/issues/548)

## 7.5.0 (April 6, 2023). Tested on Artifactory 7.55.10

IMPROVEMENTS:

* resource/artifactory_federated_*_repository: added an attribute `cleanup_on_delete`, if it's set to `true` all the federated member repositories will be deleted on `terraform destroy`. In Artifactory, if the federated repository is deleted in the UI or using the API call, federated members stayed intact to prevent losing the data. This behavior contradicts Terraform logic, when all the resources should be destroyed.
  PR: [#714](https://github.com/jfrog/terraform-provider-artifactory/pull/714)
  Issue: [#704](https://github.com/jfrog/terraform-provider-artifactory/issues/704)
* resource/artifactory_ldap_group_setting: fixed documentation.
  Issues: [#711](https://github.com/jfrog/terraform-provider-artifactory/issues/711), [#712](https://github.com/jfrog/terraform-provider-artifactory/issues/712)

## 7.4.3 (March 29, 2023). Tested on Artifactory 7.55.9

BUG FIXES:

* resource/artifactory_scoped_token: Fix not able to set `expires_in` with `0` value for non-expiring token.

PR: [#708](https://github.com/jfrog/terraform-provider-artifactory/pull/708)

## 7.4.2 (March 28, 2023). Tested on Artifactory 7.55.9

IMPROVEMENTS:

* `project_key` attribute validation for all the resources has been changed to match Artifactory requirements since 7.56.2 - the length should be between 2-32 characters.
  PR: [#707](https://github.com/jfrog/terraform-provider-artifactory/pull/707)

## 7.4.1 (March 28, 2023). Tested on Artifactory 7.55.9

IMPROVEMENTS:

* resource/artifactory_*_repository: Updates `project_environments` attribute validation. Before Artifactory 7.53.1, up to 2 values (`DEV` and `PROD`) are allowed. From 7.53.1 onward, only one value (`DEV`, `PROD`, or one of custom environment) is allowed.
PR:    [#706](https://github.com/jfrog/terraform-provider-artifactory/pull/706)
Issue: [#705](https://github.com/jfrog/terraform-provider-artifactory/issues/705)

## 7.4.0 (March 20, 2023). Tested on Artifactory 7.55.8

FEATURES:

* datasource/artifactory_federated_*_repository: Adds new data sources for all federated repository package types.
  PR:     [#693](https://github.com/jfrog/terraform-provider-artifactory/pull/693)
  Issues:
  * [#548](https://github.com/jfrog/terraform-provider-artifactory/issues/548)
  * [#692](https://github.com/jfrog/terraform-provider-artifactory/issues/692)

## 7.3.1 (March 17, 2023). Tested on Artifactory 7.55.8

BUG FIXES:

* provider: Fix panic if attribute list contain empty string (`""`) value.
PR:    [#698](https://github.com/jfrog/terraform-provider-artifactory/pull/698)
Issue: [#679](https://github.com/jfrog/terraform-provider-artifactory/issues/679)

## 7.3.0 (March 17, 2023). Tested on Artifactory 7.55.8

IMPROVEMENTS:

* resource/artifactory_push_replication and artifactory_push_replication are deprecated in favor of several new resources, listed below. Most of the attributes are not `Computed` anymore, so users can set and modify them.
* resource/artifactory_local_repository_multi_replication, artifactory_local_repository_single_replication and artifactory_remote_repository_replication were added instead of deprecated resources, listed above. Resource names reflect resource logic more clear, new attributes added. 
 PR: [#694](https://github.com/jfrog/terraform-provider-artifactory/pull/694)
 Issue: [#547](https://github.com/jfrog/terraform-provider-artifactory/issues/547)

## 7.2.1 (March 13, 2023). Tested on Artifactory 7.55.6

IMPROVEMENTS:

* resource/artifactory_scoped_token: When `expires_in` attribute is set to value that is less than Artifactory's persistency threshold then the token is created but never saved to the database. Add a warning message so users can potentially figure out why the Terraform state is invalid.
  PR:    [#691](https://github.com/jfrog/terraform-provider-artifactory/pull/691)
  Issue: [#684](https://github.com/jfrog/terraform-provider-artifactory/issues/684)

## 7.2.0 (March 6, 2023). Tested on Artifactory 7.55.6

FEATURES:

* datasource/artifactory_remote_*_repository: Adds new data sources for all remote repository package types.
  PR:    [#682](https://github.com/jfrog/terraform-provider-artifactory/pull/682)
  Issue: [#548](https://github.com/jfrog/terraform-provider-artifactory/issues/548)

## 7.1.3 (March 6, 2023). Tested on Artifactory 7.55.4

BUG FIXES:
* resource/artifactory_virtual_npm_repository: fixed import issue for `retrieval_cache_period_seconds` attribute.
 PR [#685](https://github.com/jfrog/terraform-provider-artifactory/pull/685)

## 7.1.2 (March 6, 2023). Tested on Artifactory 7.55.4

BUG FIXES:
* Changed location of data sources docs so that they render properly in the terraform registry.
  PR: [#683](https://github.com/jfrog/terraform-provider-artifactory/pull/683)
## 7.1.1 (March 2, 2023). Tested on Artifactory 7.55.2

BUG FIXES:
* resource/artifactory_remote_docker_repository, resource/artifactory_remote_helm_repository: fixed the issue when `external_dependencies_enabled` was impossible to update.
Removed constraints from `external_dependencies_patterns` attribute, now it can be set when `external_dependencies_enabled` is set to false. This is a workaround for the Artifactory API behavior, when the default value [**] is assigned instead of an empty list on the update repository call.
 PR: [#678](https://github.com/jfrog/terraform-provider-artifactory/pull/678)
 Issue: [#673](https://github.com/jfrog/terraform-provider-artifactory/issues/673)

## 7.1.0 (March 1, 2023). Tested on Artifactory 7.55.2

FEATURES:
* datasource/artifactory_local_*_repository: Adds new data sources for all local repository types.
  PR:    [#664](https://github.com/jfrog/terraform-provider-artifactory/pull/664)
  Issue: [#548](https://github.com/jfrog/terraform-provider-artifactory/issues/548)

## 7.0.2. (March 1, 2023).

BUG FIXES:
* resource/artifactory_*_repository: Fixed an issue where the new project_key default value "default" caused our provider to fail to assign and unassign repository to project.
  PR: [#674](https://github.com/jfrog/terraform-provider-artifactory/pull/674)

## 7.0.1. (February 28, 2023)

BUG FIXES:
* resource/artifactory_file: Fix `/` in artifact path being escaped. Issue: [#666](https://github.com/jfrog/terraform-provider-artifactory/issues/666) PR: [#669](https://github.com/jfrog/terraform-provider-artifactory/pull/669)

## 7.0.0. (February 27, 2023) Tested on Artifactory 7.55.0

BACKWARDS INCOMPATIBILITIES:

* resource/artifactory_*_repository: `project_key` attribute is assigned default value `default` to be compatible with Artifactory 7.50.x and above.
  It will create a state drift for Artifactory 7.49.x and below. For this reason, please use Terraform Provider Artifactory version 6.x on Artifactory 7.49.x and below.
  PR: [#668](https://github.com/jfrog/terraform-provider-artifactory/pull/668)
  Issue: [#647](https://github.com/jfrog/terraform-provider-artifactory/issues/647)

## 6.30.2 (February 27, 2023).

BUG FIXES:
* resource/artifactory_backup, resource/artifactory_ldap_group_setting, resource/artifactory_property_set, resource/artifactory_proxy, resource/artifactory_respository_layout: Fix provider erroring out instead of resetting resource ID if resource was deleted outside of Terraform. Issue: [#665](https://github.com/jfrog/terraform-provider-artifactory/issues/665) PR: [#667](https://github.com/jfrog/terraform-provider-artifactory/pull/667)

## 6.30.1 (February 24, 2023).

BUG FIXES:
* resource/artifactory_*_repository: Update `project_key` attribute validation to match Artifactory Project. PR: [#662](https://github.com/jfrog/terraform-provider-artifactory/pull/662)

## 6.30.0 (February 24, 2023).

IMPROVEMENTS:
* resource/artifactory_local_cargo_repository, resource/artifactory_remote_cargo_repository, resource/artifactory_federated_cargo_repository: Add `enable_sparse_index` attribute. PR: [#661](https://github.com/jfrog/terraform-provider-artifactory/pull/661) Issue: [#641](https://github.com/jfrog/terraform-provider-artifactory/issues/641)

## 6.29.1 (February 21, 2023).

IMPROVEMENTS:
* provider: Update `golang.org/x/net` and `golang.org/x/crypto` modules to latest version. PR: [#656](https://github.com/jfrog/terraform-provider-artifactory/pull/656) Dependabot alerts: [3](https://github.com/jfrog/terraform-provider-artifactory/security/dependabot/3), [4](https://github.com/jfrog/terraform-provider-artifactory/security/dependabot/4), [5](https://github.com/jfrog/terraform-provider-artifactory/security/dependabot/5), [6](https://github.com/jfrog/terraform-provider-artifactory/security/dependabot/6)

## 6.29.0 (February 17, 2023).

IMPROVEMENTS:
* resource/artifactory_remote_*_repository and resource/artifactory_local_*_repository: Added new attribute `cdnRedirect` for cloud users.
  PR: [#649](https://github.com/jfrog/terraform-provider-artifactory/pull/649)
  Issue: [#627](https://github.com/jfrog/terraform-provider-artifactory/issues/627)

## 6.28.1 (February 17, 2023). Tested on Artifactory 7.49.8

IMPROVEMENTS:
* data/artifactory_file: removed warning message when skipping downloading of file. Issue: [#630](https://github.com/jfrog/terraform-provider-artifactory/issues/630) PR: [#653](https://github.com/jfrog/terraform-provider-artifactory/pull/653)

## 6.28.0 (February 15, 2023). Tested on Artifactory 7.49.8

BUG FIXES:
* resource/artifactory_remote_maven_repository: renamed the attribute `metadata_retrieval_timeout_seconds` to `metadata_retrieval_timeout_secs`. This attribute can be used with any remote repository type now, not only `maven`, as it was before.
* resource/artifactory_remote_docker_repository and resource/artifactory_remote_helm_repository: fixed bug when `external_dependencies_patterns` attribute was not importable.
 PR: [#652](https://github.com/jfrog/terraform-provider-artifactory/pull/652)

## 6.27.0 (February 15, 2023). Tested on Artifactory 7.49.8

FEATURES:

* datasource/artifactory_local_*_repository: Added new data sources for some basic local repositories. Repositories included are:
  * "bower",
  * "chef",
  * "cocoapods",
  * "composer",
  * "conan",
  * "conda",
  * "cran",
  * "gems",
  * "generic",
  * "gitlfs",
  * "go",
  * "helm",
  * "npm",
  * "opkg",
  * "pub",
  * "puppet",
  * "pypi",
  * "swift",
  * "terraformbackend",
  * "vagrant"

## 6.26.1 (February 8, 2023). Tested on Artifactory 7.49.8

BUG FIXES:
* resource/artifactory_remote_*_repository: fixed bug, where remote repository password could be deleted, if it wasn't managed by the provider and `ignore_changes` was applied to that attribute.
  PR: [#634](https://github.com/jfrog/terraform-provider-artifactory/pull/643)
  Issue: [#642](https://github.com/jfrog/terraform-provider-artifactory/issues/642)

## 6.26.0 (January 31, 2023). Tested on Artifactory 7.49.6

IMPROVEMENTS:

* resource/artifactory_remote_*_repository: `propagate_query_params` attribute is removed from the common remote repository configuration. This attribute only works with Generic repo type. This change is implemented in schema V2 and migrator was added. During the migration from V1 to V2 that attribute will be removed.
  PR: [#638](https://github.com/jfrog/terraform-provider-artifactory/pull/638)
  Issue: [#635](https://github.com/jfrog/terraform-provider-artifactory/issues/635)

## 6.25.1 (January 27, 2023). Tested on Artifactory 7.49.6

BUG FIXES:

* resource/artifactory_oauth_settings: fix an issue with the import, where `oauth_provider` section couldn't be imported.
  PR: [#637](https://github.com/jfrog/terraform-provider-artifactory/pull/637)

## 6.25.0 (January 20, 2023). Tested on Artifactory 7.49.5

IMPROVEMENTS:

* Added new user data source: data.artifactory_permission_target
  Issue: [#548](https://github.com/jfrog/terraform-provider-artifactory/issues/548)
  PR: [#624](https://github.com/jfrog/terraform-provider-artifactory/pull/624/)

## 6.24.3 (January 18, 2023). Tested on Artifactory 7.49.5

IMPROVEMENTS:

* resource/artifactory_*_user: updated documentation.
  Issue: [#619](https://github.com/jfrog/terraform-provider-artifactory/issues/619)
  PR: [#629](https://github.com/jfrog/terraform-provider-artifactory/pull/629)

## 6.24.2 (January 13, 2023). Tested on Artifactory 7.49.5

BUG FIXES:

* resource/artifactory_virtual_*_repository: `omitempty` is removed from `artifactory_requests_can_retrieve_remote_artifacts` attribute, allowing users to update the value with `false` value, if it was set to `true` before.
 PR [#628](https://github.com/jfrog/terraform-provider-artifactory/pull/628)

## 6.24.1 (January 9, 2023). Tested on Artifactory 7.49.3

IMPROVEMENTS:

* resource/artifactory_*_replication: Cron expression validation replaced with verification of groups number (6 to 7). Cron verification will happen on the Artifactory API side to match UI behavior. Added more tests, documentation updated for both resources.
  Issue: [#591](https://github.com/jfrog/terraform-provider-artifactory/issues/591)
  PR: [#618](https://github.com/jfrog/terraform-provider-artifactory/pull/618)

## 6.24.0 (January 5, 2023). Tested on Artifactory 7.49.3

IMPROVEMENTS:

* Added new user data source: data.artifactory_user
  Issue: [#548](https://github.com/jfrog/terraform-provider-artifactory/issues/548)
  PR: [#611](https://github.com/jfrog/terraform-provider-artifactory/issues/611)

## 6.23.0 (January 4, 2023). Tested on Artifactory 7.49.3

IMPROVEMENTS:

* resource/artifactory_remote_*_repository: removed `Computed` from most attributes and added default values, as they appear in the UI.
  The legacy `Computed` attributes created a problem, where user can't update or remove the value of that attribute. Now, to clear the string value, an empty string could be set as an attribute value in HCL. `omitempty` is removed from most string attributes, so the user has full control and visibility of these values.
  Added a new attribute `query_params`.

* resource/artifactory_virtual_*_repository: removed unnecessary HCL tags and `omitempty` from Description, Notes and Patterns fields. Updated descriptions.

BUG FIXES:

* resource/artifactory_remote_*_repository: fixed incorrect `remote_repo_layout` assignment for all repository resources.

  Issue: [#595](https://github.com/jfrog/terraform-provider-artifactory/issues/595)
  PR: [#616](https://github.com/jfrog/terraform-provider-artifactory/pull/616)

## 6.22.3 (January 4, 2023). Tested on Artifactory 7.49.3

BUG FIXES:

* resource/artifactory_backup, resource/artifactory_ldap_group_setting, resource/artifactory_ldap_setting, resource/artifactory_property_set, resource/artifactory_proxy, resource/artifactory_repository_layout: Fix import does not update the state. Issue: [#610](https://github.com/jfrog/terraform-provider-artifactory/issues/610) PR: [#613](https://github.com/jfrog/terraform-provider-artifactory/pull/613)

NOTES:

* resource/artifactory_remote_vcs_repository: In Artifactory version 7.49.3, the attribute `max_unique_snapshots` cannot be set/updated due to an API bug.

## 6.22.2 (December 22, 2022). Tested on Artifactory 7.47.14

BUG FIXES:

* resource/artifactory_*_repository: Update `project_key` attribute validation to match Artifactory Project. PR: [#609](https://github.com/jfrog/terraform-provider-artifactory/pull/609)

## 6.22.1 (December 21, 2022). Tested on Artifactory 7.47.14

IMPROVEMENTS:

* New documentation guide for:
  * Migrating `artifactory_local_repository`, `artifactory_remote_repository`, and `artifactory_virtual_repository` to package specific repository resources.
  * Recommendation on handling user-group relationship

PR: [#608](https://github.com/jfrog/terraform-provider-artifactory/pull/608)

## 6.22.0 (December 19, 2022). Tested on Artifactory 7.47.12

IMPROVEMENTS:

* Added new group data source: data.artifactory_group
  Issue: [#548](https://github.com/jfrog/terraform-provider-artifactory/issues/548)
  PR: [#607](https://github.com/jfrog/terraform-provider-artifactory/pull/607)

## 6.21.8 (December 15, 2022). Tested on Artifactory 7.47.12

IMPROVEMENTS:

* resource/artifactory_access_token: Remove ability to import which was never supported.
* Add documentation guide for migrating access token to scoped token.

Issue: [#573](https://github.com/jfrog/terraform-provider-artifactory/issues/573) PR: [#604](https://github.com/jfrog/terraform-provider-artifactory/pull/604)

## 6.21.7 (December 14, 2022). Tested on Artifactory 7.47.12

BUG FIXES:

* resource/artifactory_remote_docker_repository: Update URL from the documentation and HCL example. PR: [#603](https://github.com/jfrog/terraform-provider-artifactory/pull/603)

## 6.21.6 (December 14, 2022). Tested on Artifactory 7.47.12

BUG FIXES:

* resource/artifactory_federated_docker_repository: Provide backward compatibility and is aliased to `artifactory_federated_docker_v2_repository` resource. Issue: [#593](https://github.com/jfrog/terraform-provider-artifactory/issues/593) PR: [#601](https://github.com/jfrog/terraform-provider-artifactory/pull/601)
* resource/artifactory_federated_docker_v1_repository, artifactory_federated_docker_v2_repository: Add missing documentation. Issue: [#593](https://github.com/jfrog/terraform-provider-artifactory/issues/593) PR: [#601](https://github.com/jfrog/terraform-provider-artifactory/pull/601)

## 6.21.5 (December 12, 2022). Tested on Artifactory 7.47.12

IMPROVEMENTS:

* resource/artifactory_anonymous_user: Update documentation and make resource limitation more prominent. Issue: [#577](https://github.com/jfrog/terraform-provider-artifactory/issues/577) PR: [#599](https://github.com/jfrog/terraform-provider-artifactory/pull/599)
* resource/artifactory_local_*_repository, resource/artifactory_remote_*_repository, resource/artifactory_virtual_*_repository:
  updated documentation for `project_environments` and `project_key` attributes. Added guide for adding repositories to the project.
  PR: [#600](https://github.com/jfrog/terraform-provider-artifactory/pull/600)

## 6.21.4 (December 9, 2022). Tested on Artifactory 7.47.12

BUG FIXES:

* resource/artifactory_federated_alpine_repository, artifactory_federated_cargo_repository, artifactory_federated_debian_repository, artifactory_federated_docker_v1_repository, artifactory_federated_docker_v2_repository, artifactory_federated_maven_repository, artifactory_federated_nuget_repository, artifactory_federated_rpm_repository, artifactory_federated_terraform_module_repository, artifactory_federated_terraform_provider_repository: Fix attributes not being updated from Artifactory during import or refresh, and therefore cause state drift.

Issue: [#593](https://github.com/jfrog/terraform-provider-artifactory/issues/593) PR: [#597](https://github.com/jfrog/terraform-provider-artifactory/pull/597)

## 6.21.3 (December 6, 2022). Tested on Artifactory 7.47.10

BUG FIXES:

* resource/artifactory_keypair:
  * Fix updating 'passphrase' does not delete and recreate key pair.
  * Fix externally deleted key pair does not trigger Terraform to recreate.

Issue: [#594](https://github.com/jfrog/terraform-provider-artifactory/issues/594) PR: [#596](https://github.com/jfrog/terraform-provider-artifactory/pull/596)

## 6.21.2 (November 30, 2022). Tested on Artifactory 7.46.11

BUG FIXES:

* resource/artifactory_scoped_token: fix token that no longer exist doesn't trigger Terraform plan recreation. Issue: [#576](https://github.com/jfrog/terraform-provider-artifactory/issues/576) PR: [#589](https://github.com/jfrog/terraform-provider-artifactory/pull/589)

## 6.21.1 (November 29, 2022). Tested on Artifactory 7.46.11

BUG FIXES:

* resource/artifactory_virtual_*_repository: removed incorrect default value for the attribute `retrieval_cache_period_seconds`, which was set to 7200 for all package types.
  Now the attribute can only be set for the package types, that supports it in the UI: Alpine, Chef, Conan, Conda, Cran, Debian, Helm and Npm.
  PR: [#590](https://github.com/jfrog/terraform-provider-artifactory/pull/590)

## 6.21.0 (November 28, 2022). Tested on Artifactory 7.46.11

IMPROVEMENTS:

* resource/artifactory_remote_conan_repository: add `force_conan_authentication` attribute to support 'force authentication'. Issue: [#578](https://github.com/jfrog/terraform-provider-artifactory/issues/578) PR: [#588](https://github.com/jfrog/terraform-provider-artifactory/pull/588)

## 6.20.2 (November 23, 2022). Tested on Artifactory 7.46.11

BUG FIXES:

* resource/artifactory_remote_vcs_repository: fix incorrect documentation. PR: [#587](https://github.com/jfrog/terraform-provider-artifactory/pull/587)

## 6.20.1 (November 21, 2022). Tested on Artifactory 7.46.11

IMPROVEMENTS:

* resource/artifactory_permission_target: Update documentation for attribute `repositories` to include values for setting any local/remote repository options. Issue: [#583](https://github.com/jfrog/terraform-provider-artifactory/issues/583)
  PR: [#584](https://github.com/jfrog/terraform-provider-artifactory/pull/584)

## 6.20.0 (November 16, 2022). Tested on Artifactory 7.46.11

FEATURES:

* resource/artifactory_proxy: add a new resource. Issue: [#562](https://github.com/jfrog/terraform-provider-artifactory/issues/562)
  PR: [#582](https://github.com/jfrog/terraform-provider-artifactory/pull/582)

## 6.19.2 (November 11, 2022). Tested on Artifactory 7.46.11

BUG FIXES:

* resources/artifactory_keypair: add `passphrase` attribute to the JSON body. No API errors in Artifactory 7.41.13 and up. Issue: [#574](https://github.com/jfrog/terraform-provider-artifactory/issues/574)
  PR: [#581](https://github.com/jfrog/terraform-provider-artifactory/pull/581)

## 6.19.1 (November 11, 2022). Tested on Artifactory 7.46.11

IMPROVEMENTS:

* resources/artifactory_scoped_token: Add `Sensitive: true` to `access_token` and `refresh_token` attributes to ensure the values are handled correctly.

## 6.19.0 (October 25, 2022). Tested on Artifactory 7.46.10

IMPROVEMENTS:

* resource/artifactory_virtual_docker_repository: added new attribute `resolve_docker_tags_by_timestamp`. Issue: [#563](https://github.com/jfrog/terraform-provider-artifactory/issues/563)
  PR: [#PR](https://github.com/jfrog/terraform-provider-artifactory/pull/569)
* resource/artifactory_backup: added a format note to the documentation. Issue: [#564](https://github.com/jfrog/terraform-provider-artifactory/issues/564)

## 6.18.0 (October 21, 2022). Tested on Artifactory 7.46.6

IMPROVEMENTS:

* resource/artifactory_remote_nuget_repository: added new attribute `symbol_server_url`. Issue: [#549](https://github.com/jfrog/terraform-provider-artifactory/issues/549)
  PR: [#567](https://github.com/jfrog/terraform-provider-artifactory/pull/567)

## 6.17.1 (October 21, 2022)

BUG FIXES:

* Update documentation to change incorrect repository type reference 'gem' to correct type 'gems'. Issue: [#541](https://github.com/jfrog/terraform-provider-artifactory/issues/541) PR: [#566](https://github.com/jfrog/terraform-provider-artifactory/pull/566)

## 6.17.0 (October 21, 2022). Tested on Artifactory 7.46.6

IMPROVEMENTS:

* resource/artifactory_federated_swift_repository: added new resource. Issue: [#540](https://github.com/jfrog/terraform-provider-artifactory/issues/540)
  PR: [#565](https://github.com/jfrog/terraform-provider-artifactory/pull/565)

## 6.16.4 (October 17, 2022). Tested on Artifactory 7.46.6

BUG FIXES:

* resource/artifactory_remote_*_repository: removed condition to update certain fields (like `xray_index`) only if they got changed in the HCL,
  which lead to assigning the default values to these fields. Issue: [#557](https://github.com/jfrog/terraform-provider-artifactory/issues/557)
  PR: [#561](https://github.com/jfrog/terraform-provider-artifactory/pull/561)

## 6.16.3 (October 12, 2022). Tested on Artifactory 7.46.3

DEPRECATION:

* resource/artifactory_api_key: added deprecation notice. The API key support will be removed in upcoming versions of Artifactory.

## 6.16.2 (October 11, 2022)

IMPROVEMENTS:

* Update documentation to distinguish resources that are not supported by JFrog SaaS environment. Issue: [#550](https://github.com/jfrog/terraform-provider-artifactory/issues/550) PR: [#551](https://github.com/jfrog/terraform-provider-artifactory/pull/551)
* Remove `make doc` command from make file. Issue: [#552](https://github.com/jfrog/terraform-provider-artifactory/issues/552) PR: [#555](https://github.com/jfrog/terraform-provider-artifactory/pull/555)

## 6.16.1 (October 10, 2022). Tested on Artifactory 7.41.13

IMPROVEMENTS:

* resource/artifactory_remote_*_repository: attribute 'remote_repo_layout_ref' is deprecated. Issue: [#542](https://github.com/jfrog/terraform-provider-artifactory/issues/542)
  PR: [#553](https://github.com/jfrog/terraform-provider-artifactory/pull/553)

NOTE: 'remote_repo_layout_ref' will be removed on the next major release.

## 6.16.0 (September 27, 2022). Tested on Artifactory 7.41.13

FEATURES:

* resource/artifactory_property_set: add a new resource. Issue: [#522](https://github.com/jfrog/terraform-provider-artifactory/issues/522)
  PR: [#546](https://github.com/jfrog/terraform-provider-artifactory/pull/546)

## 6.15.1 (September 14, 2022). Tested on Artifactory 7.41.12

IMPROVEMENTS:

* resource/artifactory_*_repository: Use projects API to assign/unassign to project when project_key is set/unset for existing repo. Issue: [#329](https://github.com/jfrog/terraform-provider-artifactory/issues/329) PR: [#537](https://github.com/jfrog/terraform-provider-artifactory/pull/537)

BUG FIXES:

* resource/artifactory_repository_layout: Add missing documentation. PR: [#538](https://github.com/jfrog/terraform-provider-artifactory/pull/538)

## 6.15.0 (August 31, 2022)

IMPROVEMENTS:

* resource/artifactory_remote_*_repostiory: Add attribute `download_direct`. PR: [#528](https://github.com/jfrog/terraform-provider-artifactory/pull/528)

## 6.14.1 (August 26, 2022). Tested on Artifactory 7.41.7

BUG FIXES:

* resource/artifactory_scoped_token: Add missing `refresh_token` attribute for output. Issue: [#531](https://github.com/jfrog/terraform-provider-artifactory/issues/531) PR: [#533](https://github.com/jfrog/terraform-provider-artifactory/pull/533).

## 6.14.0 (August 26, 2022). Tested on Artifactory 7.41.7

FEATURES:

* **New Resource:** `artifactory_repository_layout` Issue: [#503](https://github.com/jfrog/terraform-provider-artifactory/issues/503) PR: [#532](https://github.com/jfrog/terraform-provider-artifactory/pull/532).

## 6.13.0 (August 24, 2022). Tested on Artifactory 7.41.7

IMPROVEMENTS:

* resource/artifactory_backup: Add attributes `verify_disk_space` and `export_mission_control`. Issue: [#516](https://github.com/jfrog/terraform-provider-artifactory/issues/516) PR: [#530](https://github.com/jfrog/terraform-provider-artifactory/pull/530).

## 6.12.1 (August 23, 2022). Tested on Artifactory 7.41.7

BUG FIXES:

* resource/artifactory_remote_*_repository: Fix unable to reset `excludes_pattern` attribute using empty string. PR: [#527](https://github.com/jfrog/terraform-provider-artifactory/pull/527).

## 6.12.0 (August 17, 2022). Tested on Artifactory 7.41.7

IMPROVEMENTS:

* resource/artifactory_remote_maven_repository: Add attribute `metadata_retrieval_timeout_seconds`. Issue: [#509](https://github.com/jfrog/terraform-provider-artifactory/issues/509) PR: [#525](https://github.com/jfrog/terraform-provider-artifactory/pull/525).

## 6.11.3 (August 9, 2022). Tested on Artifactory 7.41.7

BUG FIXES:

* resource/artifactory_*_repository: Add support for hyphen character in `project_key` attribute. PR: [#524](https://github.com/jfrog/terraform-provider-artifactory/pull/524).

## 6.11.2 (July 28, 2022). Tested on Artifactory 7.41.6

IMPROVEMENTS:

* resource/artifactory_push_replication: Improve sample HCL in documentation. PR: [#519](https://github.com/jfrog/terraform-provider-artifactory/pull/519).
* resourec/artifactory_virtual_maven_repository: Improve sample HCL in documentation. PR: [#519](https://github.com/jfrog/terraform-provider-artifactory/pull/519)
* resource/artifactory_user: Fix inaccurate descriptions for attributes `profile_updatable` and `disable_ui_access`. PR: [#517](https://github.com/jfrog/terraform-provider-artifactory/pull/517). Issue: [#518](https://github.com/jfrog/terraform-provider-artifactory/issues/518)

## 6.11.1 (July 20, 2022). Tested on Artifactory 7.41.4

BUG FIXES:

* resource/artifactory_saml_settings: Fix attribute `no_auto_user_creation` has opposite result. PR: [#512](https://github.com/jfrog/terraform-provider-artifactory/pull/512). Issue: [#500](https://github.com/jfrog/terraform-provider-artifactory/issues/500)
* resourec/artifactory_api_key: Fix failed acceptance test. PR: [#511](https://github.com/jfrog/terraform-provider-artifactory/pull/511)

## 6.11.0 (July 8, 2022)

IMPROVEMENTS:

* Support for swift repo [#497](https://github.com/jfrog/terraform-provider-artifactory/pull/505). Issue: [#496](https://github.com/jfrog/terraform-provider-artifactory/issues/489)

DOCUMENTATION:

* Added `api_key` deprecation message.

## 6.10.1 (July 1, 2022)

BUG FIXES:

* Hack around [weird terraform bug](https://discuss.hashicorp.com/t/using-typeset-in-provider-always-adds-an-empty-element-on-update/18566/2) dealing with sets. PR: [#481](https://github.com/jfrog/terraform-provider-artifactory/pull/496). Issue: [#496](https://github.com/jfrog/terraform-provider-artifactory/issues/476)
* provider: Fix hardcoded HTTP user-agent string. PR: [#497](https://github.com/jfrog/terraform-provider-artifactory/pull/497)

## 6.10.0 (June 28, 2022)

IMPROVEMENTS:

* resource/artifactory_permission_target: Add support for `distribute` permission for `release_bundle`. PR: [#490](https://github.com/jfrog/terraform-provider-artifactory/pull/490)

## 6.9.6 (June 27, 2022). Tested on Artifactory 7.38.10

REFACTOR:

* Updated docs for `local_maven_repository` PR: [#493](https://github.com/jfrog/terraform-provider-artifactory/pull/493). Issue: [#480](https://github.com/jfrog/terraform-provider-artifactory/issues/488)

## 6.9.5 (June 27, 2022). Tested on Artifactory 7.38.10

REFACTOR:

* Moved some functionality to shared
* Fixed tests

## 6.9.4 (June 21, 2022). Tested on Artifactory 7.38.10

REFACTOR:

* Remove redundant shared code to shared module and bump dependency.
* Moved some other sharable code to shared module

## 6.9.3 (June 10, 2022). Tested on Artifactory 7.38.10

BUG FIXES:

* resource/artifactory_file: Check for file existence before verifying checksum. PR: [#481](https://github.com/jfrog/terraform-provider-artifactory/pull/481). Issue: [#480](https://github.com/jfrog/terraform-provider-artifactory/issues/480)

## 6.9.2 (June 7, 2022). Tested on Artifactory 7.38.10

BUG FIXES:

* resource/artifactory_scoped_token:
  * Expand `audiences` validation to include all valid JFrog service types. PR: [#477](https://github.com/jfrog/terraform-provider-artifactory/pull/477). Issue: [#475](https://github.com/jfrog/terraform-provider-artifactory/issues/475)
  * Fix incorrect validation for `applied-permissions/groups` scope. PR: [#477](https://github.com/jfrog/terraform-provider-artifactory/pull/477). Issue: [#478](https://github.com/jfrog/terraform-provider-artifactory/issues/478)

## 6.9.1 (June 3, 2022). Tested on Artifactory 7.38.10

BUG FIXES:

* resource/artifactory_virtual_npm_repository: Add missing attributes `external_dependencies_enabled`, `external_dependencies_patterns`, and `external_dependencies_remote_repo`. PR: [#473](https://github.com/jfrog/terraform-provider-artifactory/pull/473). Issue: [#463](https://github.com/jfrog/terraform-provider-artifactory/issues/463)

## 6.9.0 (May 24, 2022). Tested on Artifactory 7.38.10

FEATURES:

* Added new resources to support Terraform repositories.
  * Local: Terraform Module (`resource/artifactory_local_terraform_module_repository`).
    Terraform Provider (`resource/artifactory_local_terraform_provider_repository`) and Terraform Backend (`resource\artifactory_local_terraformbackend_repository`).
  * Remote: Terraform Repository (`resource/artifactory_remote_terraform_repository`).
  * Virtual: Terraform Repository (`resource/artifactory_virtual_terraform_repository`).
  * Federated: Terraform Module (`resource/artifactory_federated_terraform_module_repository`), Terraform Provider (`resource/artifactory_federated_terraform_provider_repository`).

    Issue [#450](https://github.com/jfrog/terraform-provider-artifactory/issues/450)
    PR: [#464](https://github.com/jfrog/terraform-provider-artifactory/pull/464).

## 6.8.2 (June 2, 2022). Tested on Artifactory 7.38.10

BUG FIXES:

* resource/artifactory_local_maven_repository, resource/artifactory_local_gradle_repository, resource/artifactory_local_sbt_repository, resource/artifactory_local_ivy_repositor: Fix validation for attribute `checksum_policy_type`. Previously it accepts `generated-checksums`. Now it accepts `server-generated-checksums`. Same applies to the corresponding federated repository resources. PR: [#471](https://github.com/jfrog/terraform-provider-artifactory/pull/471). Issue [#470](https://github.com/jfrog/terraform-provider-artifactory/issues/470)

## 6.8.1 (May 31, 2022). Tested on Artifactory 7.38.10

ENHANCEMENTS:

* resource/artifactory_file: Add debugging loggings to aid investigate issue. PR: [#466](https://github.com/jfrog/terraform-provider-artifactory/pull/466) Issue: [#441](https://github.com/jfrog/terraform-provider-artifactory/issues/441)

## 6.8.0 (May 31, 2022). Tested on Artifactory 7.38.10

FEATURES:

* resource/artifactory_scoped_token: New resource for Artifactory scoped token. PR: [#465](https://github.com/jfrog/terraform-provider-artifactory/pull/465). Issue [#451](https://github.com/jfrog/terraform-provider-artifactory/issues/451)

## 6.7.3 (May 27, 2022). Tested on Artifactory 7.38.10

IMPROVEMENTS:

* Upgrade `gopkg.in/yaml.v3` to v3.0.0 for [CVE-2022-28948](https://nvd.nist.gov/vuln/detail/CVE-2022-28948) PR [#467](https://github.com/jfrog/terraform-provider-artifactory/pull/467)

## 6.7.2 (May 13, 2022). Tested on Artifactory 7.38.8

IMPROVEMENTS:

* resource/artifactory_pull_replication.go and resource/artifactory_push_replication.go: Add new attribute `check_binary_existence_in_filestore`.
  PR: [#460](https://github.com/jfrog/terraform-provider-artifactory/pull/460).
  Issue [#434](https://github.com/jfrog/terraform-provider-artifactory/issues/434)

## 6.7.1 (May 13, 2022). Tested on Artifactory 7.38.8

BUG FIXES:

* resource/artifactory_federated_*_repository: Fix attributes from corresponding local repository were not used. PR: [#458](https://github.com/jfrog/terraform-provider-artifactory/pull/458). Issue [#431](https://github.com/jfrog/terraform-provider-artifactory/issues/431)

## 6.7.0 (May 12, 2022). Tested on Artifactory 7.38.8

IMPROVEMENTS:

* resource/artifactory_*_webhook: Add support for multiple outlets (handlers) of the webhook. Existing attributes (`url`, `secret`, `proxy`, and `custom_http_headers`) will be automatically migrated to be the first handler.

To migrate to new webhook schema with multiple handlers:
- Update your HCL and copy the attributes (`url`, `secret`, `proxy`, and `custom_http_headers`) into a `handler` block (See `sample.tf` for full examples)
- Execute `terraform apply -refresh-only` to update the Terraform state

Issue [#439](https://github.com/jfrog/terraform-provider-artifactory/issues/439) PR: [#453](https://github.com/jfrog/terraform-provider-artifactory/pull/453).

BUG FIXES:

* resource/artifactory_permission_target: Fix not working `release_bundle` attribute PR: [#454](https://github.com/jfrog/terraform-provider-artifactory/pull/454). Issue [#449](https://github.com/jfrog/terraform-provider-artifactory/issues/449)

## 6.6.2 (May 11, 2022). Tested on Artifactory 7.38.8

BUG FIXES:

* provider: Fix license checking only works with 'Enterprise' license type. PR: [#456](https://github.com/jfrog/terraform-provider-artifactory/pull/456). Issue [#455](https://github.com/jfrog/terraform-provider-artifactory/issues/455)

## 6.6.1 (May 5, 2022). Tested on Artifactory 7.37.16

BUG FIXES:

* resource/artifactory_federated_*_repository: Use correct 'base' schema from local repository. PR: [#443](https://github.com/jfrog/terraform-provider-artifactory/pull/443). Issue [#431](https://github.com/jfrog/terraform-provider-artifactory/issues/431)

## 6.6.0 (Apr 29, 2022). Tested on Artifactory 7.37.15

IMPROVEMENTS:

* resource/artifactory_group: Add `external_id` attribute to support Azure AD group. PR: [#437](https://github.com/jfrog/terraform-provider-artifactory/pull/437). Issue [#429](https://github.com/jfrog/terraform-provider-artifactory/issues/429)

## 6.5.3 (Apr 27, 2022). Tested on Artifactory 7.37.15

IMPROVEMENTS:

* reorganizing documentation, adding missing documentation links, fixing formatting. No changes in the functionality.
  PR: [GH-435](https://github.com/jfrog/terraform-provider-artifactory/pull/435). Issues [#422](https://github.com/jfrog/terraform-provider-artifactory/issues/422) and [#398](https://github.com/jfrog/terraform-provider-artifactory/issues/398)

## 6.5.2 (Apr 25, 2022). Tested on Artifactory 7.37.14

IMPROVEMENTS:

* resource/artifactory_artifact_webhook: Added 'cached' event type for Artifact webhook. PR: [GH-430](https://github.com/jfrog/terraform-provider-artifactory/pull/430).

## 6.5.1 (Apr 20, 2022). Tested on Artifactory 7.37.14

BUG FIXES:

* provider:  Setting the right default value for 'access_token' attribute. PR: [GH-426](https://github.com/jfrog/terraform-provider-artifactory/pull/426). Issue [#425](https://github.com/jfrog/terraform-provider-artifactory/issues/425)

## 6.5.0 (Apr 19, 2022). Tested on Artifactory 7.37.14

IMPROVEMENTS:

* Resources added for Pub package type of Local Repository
* Resources added for Pub package type of Remote Repository
* Resources added for Pub package type of Virtual Repository
* Acceptance test case enhanced with Client TLS Certificate

PR: [GH-421](https://github.com/jfrog/terraform-provider-artifactory/pull/421)

## 6.4.1 (Apr 18, 2022). Tested on Artifactory 7.37.14

IMPROVEMENTS:

* provider: Support `JFROG_ACCESS_TOKEN` environment variable source for 'access_token' attribute. [GH-418]

## 6.4.0 (Apr 15, 2022). Tested on Artifactory 7.37.13

FEATURES:

* Added new `artifactory_unmanaged_user` resource which is an alias of existing `artifactory_user`.
* Added new `artifactory_managed_user` resource with `password` attribute being required and no automatic password generation.
* Added new `artifactory_anonymous_user` resource which allows importing of Artifactory 'anonymous' user into Terraform state.

[GH-396]

## 6.3.0 (Apr 15, 2022). Tested on Artifactory 7.37.13

IMPROVEMENTS:

* resource/artifactory_permission_targets: Add deprecation message [GH-413]
* Removed dependency on `jfrog-client-go` package [GH-413]

NOTES:

* Resource `artifactory_permission_targets` is deprecated and will be removed in the next major release. Resource `artifactory_permission_target` (singular) has an identical schema which will allow straightforward migration.

## 6.2.0 (Apr 15, 2022). Tested on Artifactory 7.35.2

BUG FIXES:

* resource/artifactory_pull_replication: Make `password` attribute configurable. `url`, `username`, and `password` attributes must be set together when use with remote repository. [GH-411]
* resource/artifactory_push_replication: Make `password` attribute configurable. `url`, `username`, and `password` attributes are now required to match Artifactory API requirements [GH-411]

## 6.1.3 (Apr 12, 2022)

BUG FIXES:

* resource/artifactory_user: Fix to persist changes to groups [GH-406]

## 6.1.2 (Apr 11, 2022)

IMPROVEMENTS:

* Documentation changes for `artifactory_keypair` resource [GH-402]

## 6.1.1 (Apr 11, 2022)

BUG FIXES:

* resource/artifactory_push_replication: unable to update resource after creation [GH-400]

## 6.1.0 (Apr 11, 2022)

IMPROVEMENTS:

* Added gpg keypair attributes for `artifactory_local_rpm_repository` resource [GH-397]

## 6.0.1 (Apr 7, 2022)

IMPROVEMENTS:

* Added VCS remote repository resource - `artifactory_remote_vcs_repository` [GH-394]

## 6.0.0 (Apr 6, 2022)

BREAKING CHANGES:

* `artifactory_local_repository`, `artifactory_remote_repository` and `artifactory_virtual_repository` were removed from the provider. Please use resources with package-specific names, like `artifactory_local_cargo_repository` [GH-380]

## 5.0.0 (Apr 6, 2022)

BREAKING CHANGE:

* resource/artifactory_user: Attribute `password` is optional again. If it is omitted in the HCL, a random password is generated automatically for Artifactory user. This password is not stored in the Terraform state file and thus will not trigger a state drift. [GH-390]

## 4.0.2 (Apr 6, 2022)

BUG FIXES:

* Fix typos in `artifactory_federated_*_repository` resources documentation. [GH-391]

## 4.0.1 (Apr 4, 2022)

BUG FIXES:

* Fix remote repos' `password` attribute always being updated after initial `terraform apply` [GH-385]

## 4.0.0 (Mar 31, 2022)

BREAKING CHANGE:

* Basic authentication with username and password is removed from the provider. [GH-344]

## 3.1.4 (Mar 31, 2022)

BUG FIXES:

* Fix blank password getting sent to Artifactory when updating other attributes of `artifactory_user` resource. [GH-383]

## 3.1.3 (Mar 31, 2022)

IMPROVEMENTS:

* Documentation improved for `artifactory_general_security` resource. [GH-367]

## 3.1.2 (Mar 31, 2022)

BUG FIXES:

* Fix proxy getting unset after modifying existing artifactory_remote_*_repository resources. [GH-381]

## 3.1.1 (Mar 30, 2022)

BUG FIXES:

* resource/artifactory_local_docker_v2_repository: Fix `max_unique_tags` with value 0 being ignored. [GH-376]

## 3.1.0 (Mar 29, 2022)

FEATURES:

* **New Resources:** Added following local repository resources in new implementation. [GH-378]
  * "artifactory_local_cargo_repository"
  * "artifactory_local_conda_repository"

## 3.0.2 (Mar 29, 2022)

IMPROVEMENTS:

* Update module path to `/v3` in `go.mod` and `main.go` [GH-374]

## 3.0.1 (Mar 28, 2022)

BUG FIXES:

* Fix retrieval_cache_period_seconds to be set to 0 for artifactory_remote_*_repository resources. [GH-373]

## 3.0.0 (Mar 28, 2022)

BREAKING CHANGES:

* Resources `artifactory_xray_policy` and `artifactory_xray_watch` have been removed [GH-315]

## 2.25.0 (Mar 21, 2022)

FEATURES:

* **New Resources:** Added following virtual repository resources in new implementation. [GH-365]
  * "artifactory_virtual_alpine_repository"
  * "artifactory_virtual_bower_repository"
  * "artifactory_virtual_chef_repository"
  * "artifactory_virtual_conda_repository"
  * "artifactory_virtual_composer_repository"
  * "artifactory_virtual_cran_repository"
  * "artifactory_virtual_debian_repository"
  * "artifactory_virtual_docker_repository"
  * "artifactory_virtual_gems_repository"
  * "artifactory_virtual_gitlfs_repository"
  * "artifactory_virtual_gradle_repository"
  * "artifactory_virtual_ivy_repository"
  * "artifactory_virtual_npm_repository"
  * "artifactory_virtual_nuget_repository"
  * "artifactory_virtual_p2_repository"
  * "artifactory_virtual_puppet_repository"
  * "artifactory_virtual_pypi_repository"
  * "artifactory_virtual_sbt_repository"

## 2.24.0 (Mar 18, 2022)

FEATURES:

* **New Resources:** Added following remote repository resources in new implementation. [GH-364]
  * "artifactory_remote_alpine_repository"
  * "artifactory_remote_bower_repository"
  * "artifactory_remote_chef_repository"
  * "artifactory_remote_cocoapods_repository"
  * "artifactory_remote_conda_repository"
  * "artifactory_remote_conan_repository"
  * "artifactory_remote_composer_repository"
  * "artifactory_remote_cran_repository"
  * "artifactory_remote_debian_repository"
  * "artifactory_remote_gems_repository"
  * "artifactory_remote_go_repository"
  * "artifactory_remote_generic_repository"
  * "artifactory_remote_gitlfs_repository"
  * "artifactory_remote_opkg_repository"
  * "artifactory_remote_p2_repository"
  * "artifactory_remote_puppet_repository"
  * "artifactory_remote_rpm_repository"
  * "artifactory_remote_nuget_repository"

## 2.23.2 (Mar 17, 2022)

IMPROVEMENTS:

* Datasource `datasource_artifactory_file`, added a parameter `path_is_aliased`,
  assumes that the path supplied is an alias for the most recent version of the artifact and doesn't try to resolve it to a specific, timestamped, artifact

## 2.23.1 (Mar 15, 2022)

IMPROVEMENTS:

* resource/artifactory_remote_docker_repository: Setting default value '**' for external_dependencies_patterns field. [GH-363]
* resource/artifactory_remote_helm_repository: Setting default value '**' for external_dependencies_patterns field. [GH-363]

## 2.23.0 (Mar 11, 2022)

FEATURES:

* **New Resources:** Added following local and remote repository resources in new implementation. [GH-360]
  * "artifactory_local_sbt_repository"
  * "artifactory_local_ivy_repository"
  * "artifactory_remote_sbt_repository"
  * "artifactory_remote_ivy_repository"

## 2.22.3 (Mar 10, 2022)

BUG FIXES:

* Conditional file download depending on `force_overwrite` value of data source `artifactory_file`. [GH-352]

## 2.22.2 (Mar 10, 2022)

BUG FIXES:

* resource/artifactory_ldap_setting: Made user_dn_pattern attribute optional. [GH-356]

## 2.22.1 (Mar 8, 2022)

IMPROVEMENTS:

* Make repository layout to correct default value as per package type, provided the `repo_layout_ref` attribute is not supplied explicitly in the resource. [GH-335]

## 2.22.0 (Mar 8, 2022)

FEATURES:

* resource/artifactory_push_replication: Add support for specifying proxy. [GH-337]
* resource/artifactory_replication_config: Add support for specifying proxy. [GH-337]
* resource/artifactory_single_replication: Add support for specifying proxy. [GH-337]

## 2.21.0 (Mar 3, 2022)

FEATURES:

* **New Resources:** Added following remote repository resources. [GH-343]
  * "artifactory_remote_maven_repository"
  * "artifactory_remote_gradle_repository"

## 2.20.4 (Feb 28, 2022)

IMPROVEMENTS:

* resource/artifactory_remote_docker_repository: Added list_remote_folder_items attribute to resource_artifactory_remote_docker_repository. [GH-338]
* resource/artifactory_remote_cargo_repository: Added list_remote_folder_items attribute to resource_artifactory_remote_cargo_repository. [GH-338]
* resource/artifactory_remote_helm_repository: Added list_remote_folder_items attribute to resource_artifactory_remote_helm_repository. [GH-338]
* resource/artifactory_remote_pypi_repository: Added list_remote_folder_items attribute to resource_artifactory_remote_pypi_repository. [GH-338]

## 2.20.3 (Feb 25, 2022)

IMPROVEMENTS:

* Add previously missing repository resource attributes to documentation [GH-332]

## 2.20.2 (Feb 25, 2022)

IMPROVEMENTS:

* resource/artifactory_backup: Added support for system backup configuration. [GH-331]

## 2.20.1 (Feb 24, 2022)

IMPROVEMENTS:

* Make `xray_index` attribute for local/remote/federated repository resources settable by users [GH-330]
* Add documentation for `xray_index`  [GH-330]

## 2.20.0 (Feb 20, 2022)

FEATURES:

* resource/artifactory_virtual_helm_repository: New resource for Helm repository type with namespaces support [GH-322]

## 2.19.1 (Feb 16, 2022)

IMPROVEMENTS:

* Add a test and update the sample TF for `artifactory_remote_pypi_repository` [GH-321]

## 2.19.0 (Feb 16, 2022)

IMPROVEMENTS:

* Add `project_key` and `project_environments` to local, remote, virtual, and federated repository resources to support Artifactory Projects [GH-320]

## 2.18.1 (Feb 14, 2022)

BUG FIXES:

* resource/artifactory_keypair: Fix key pair not being stored in Terraform state correctly. [GH-317]

## 2.18.0 (Feb 14, 2022)

FEATURES:

* **New Resources:** Webhook resources [GH-313]
  * `artifactory_artifact_webhook`
  * `artifactory_artifact_property_webhook`
  * `artifactory_docker_webhook`
  * `artifactory_build_webhook`
  * `artifactory_release_bundle_webhook`
  * `artifactory_distribution_webhook`
  * `artifactory_artifactory_release_bundle_webhook`

## 2.17.0 (Feb 12, 2022)

IMPROVEMENTS:

* resource/resource_artifactory_remote_pypi_repository: Added support for pypi remote repository with fix for priority_resolution attribute. [GH-316]

## 2.16.2 (Feb 10, 2022)

BUG FIXES:

* resource/artifactory_single_replication_config: Fix for error when repository got externally removed, but replication resource configured. [GH-312]

## 2.16.1 (Feb 7, 2022)

BUG FIXES:

* resource/artifactory_remote_repository: Fix failing test for `proxy` attribute [GH-311]

## 2.16.0 (Feb 4, 2022)

IMPROVEMENTS:

* resource/artifactory_group: Added support for manager roles in artifactory_group resource [GH-308]

## 2.15.2 (Feb 4, 2022)

BUG FIXES:

* resource/artifactory_remote_repository: Fix unable to reset `proxy` attribute [GH-307]

## 2.15.1 (Feb 4, 2022)

BUG FIXES:

* resource/artifactory_xray_watch: Fix incorrect usage of variable reference with Resty `.SetBody()` in `create` and `update` funcs [GH-306]

## 2.15.0 (Feb 3, 2022)

FEATURES:

* **New Resource:** `artifactory_virtual_rpm_repository` with support for `primary_keypair_ref` and `secondary_keypair_ref` and [GH-303]

## 2.14.0 (Feb 3, 2022)

FEATURES:

* Added following smart remote repo attributes for npm, cargo, docker and helm remote repository resources [GH-305].
  * "statistics_enabled"
  * "properties_enabled"
  * "source_origin_absence_detection"

## 2.13.1 (Feb 2, 2022)

IMPROVEMENTS:

* Add missing documentations for Federated repo resources [GH-304]
* Add additional repo types for Federated repo resources [GH-304]

## 2.13.0 (Feb 1, 2022)

FEATURES:

* **New Resources:** `artifactory_federated_x_repository` where `x` is one of the following [GH-296]:
  * "bower"
  * "chef"
  * "cocoapods"
  * "composer"
  * "conan"
  * "cran"
  * "gems"
  * "generic"
  * "gitlfs"
  * "go"
  * "helm"
  * "ivy"
  * "npm"
  * "opkg"
  * "puppet"
  * "pypi"
  * "sbt"
  * "vagrant"
