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

    PR: [#464](https://github.com/jfrog/terraform-provider-artifactory/pull/464).
    Issue [#450](https://github.com/jfrog/terraform-provider-artifactory/issues/450)

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

PR: [#453](https://github.com/jfrog/terraform-provider-artifactory/pull/453). Issue [#439](https://github.com/jfrog/terraform-provider-artifactory/issues/439)

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
