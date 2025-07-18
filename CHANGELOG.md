### 12.9.7 (July 17, 2025). Tested on Artifactory 7.111.12 with Terraform 1.12.2 and OpenTofu 1.10.3

IMPROVEMENTS:

* resource/artifactory_local_repository_single_replication: Introduced a new attribute disable_proxy for single replication. This allows users to explicitly disable the use of a proxy when configuring single replication in the Artifactory Terraform provider. Issue: [#1255](https://github.com/jfrog/terraform-provider-artifactory/issues/1255). PR: [#1284](https://github.com/jfrog/terraform-provider-artifactory/pull/1284)
* resource/artifactory_\*\_repository: Remove static set size validator for `project_environments`. Issue: [#1276](https://github.com/jfrog/terraform-provider-artifactory/issues/1276) PR: [#1283](https://github.com/jfrog/terraform-provider-artifactory/pull/1283)

BUG FIXES:

* resource/artifactory_\*\_repository: Fix prevent content_synchronisation block state drift in remote repositories PR: [#1295](https://github.com/jfrog/terraform-provider-artifactory/pull/1295)
* resource/artifactory_package_cleanup_policy: Fix validation for time-based and version-based and property-based conditions, added included_properties and excluded_properties attributes. Issue: [#1285](https://github.com/jfrog/terraform-provider-artifactory/issues/1285). Issue: [#1289](https://github.com/jfrog/terraform-provider-artifactory/issues/1289), Issue: [#1290](https://github.com/jfrog/terraform-provider-artifactory/issues/1290) PR: [#1293](https://github.com/jfrog/terraform-provider-artifactory/pull/1293)

## 12.9.6 (June 17, 2025). Tested on Artifactory 7.111.10 with Terraform 1.12.2 and OpenTofu 1.9.1

NOTES:

If you are upgrading from a version earlier than 12.8.3, please upgrade to 12.8.3 first. Then proceed to the latest version to avoid state drift in the content_synchronisation attribute of remote repositories.

IMPROVEMENTS:

* GNUmakefile : Enhanced ARM64 Build Process with Dynamic GOARM64 Detection. PR: [#1282](https://github.com/jfrog/terraform-provider-artifactory/pull/1282)

BUG FIXES:

* resource/artifactory_\*\_repository: Fix state drift handling for content_synchronisation attribute in remote repo's. Issue: [#1250](https://github.com/jfrog/terraform-provider-artifactory/issues/1250). PR: [#1274](https://github.com/jfrog/terraform-provider-artifactory/pull/1274)

## 12.9.5 (June 3, 2025). Tested on Artifactory 7.111.8 with Terraform 1.12.0 and OpenTofu 1.9.1

BUG FIXES:

* resource/artifactory_package_cleanup_policy : Fix the cron expression validation failure for "0 0 2 ? * SAT" (run every Saturday at 2am). Issue: [#1247](https://github.com/jfrog/terraform-provider-artifactory/issues/1247). PR: [#1272](https://github.com/jfrog/terraform-provider-artifactory/pull/1272)

## 12.9.4 (May 19, 2025). Tested on Artifactory 7.111.8 with Terraform 1.12.0 and OpenTofu 1.9.1

FEATURES:

**New Resource:** `artifactory_release_bundle_v2_cleanup_policy` to support release bundle v2 cleanup policy. PR: [#1161](https://github.com/jfrog/terraform-provider-artifactory/pull/1266)

## 12.9.3 (May 15, 2025). Tested on Artifactory 7.111.4 with Terraform 1.11.3 and OpenTofu 1.9.0

BUG FIXES:

* resource/artifactory_property_set : Fix to remove the enforcement of artificial requirements on predefined_value. It is now only mandatory when closed_predefined_values or multiple_choice is set to true. Issue: [#1214](https://github.com/jfrog/terraform-provider-artifactory/issues/1214) PR: [#1240](https://github.com/jfrog/terraform-provider-artifactory/pull/1240)
* resource/resource_artifactory_scoped_token : Fix #Validation of scope when creating tokens doesn't include all valid options. Issue: [#1235](https://github.com/jfrog/terraform-provider-artifactory/issues/1235) PR: [#1241](https://github.com/jfrog/terraform-provider-artifactory/pull/1241)
* resource/artifactory_*_repository: Fix to enable multiple project environments for repositories in Artifactory 7.107.1 and later.

## 12.9.2 (April 2, 2025). Tested on Artifactory 7.104.14 with Terraform 1.11.3 and OpenTofu 1.9.0

BUG FIXES:

* resource/artifactory_remote_docker_repository,resource/artifactory_remote_helmoci_repository,resource/artifactory_remote_oci_repository : Fix artifactory_remote_docker_repository shows constant diff on external_dependencies_patterns. Issue: [#1217](https://github.com/jfrog/terraform-provider-artifactory/issues/1217) PR: [#1237](https://github.com/jfrog/terraform-provider-artifactory/pull/1237)

## 12.9.1 (Februry 25, 2025). Tested on Artifactory 7.104.9 with Terraform 1.10.5 and OpenTofu 1.9.0

BUG FIXES:

* resource/artifactory_\*\_repository: Improve state drift handling. Issue: [#1200](https://github.com/jfrog/terraform-provider-artifactory/issues/1200) PR: [#1212](https://github.com/jfrog/terraform-provider-artifactory/pull/1212)
* resource/artifactory_\*\_webhook: Improve state drift handling. PR: [#1212](https://github.com/jfrog/terraform-provider-artifactory/pull/1212)

## 12.9.0 (Februry 20, 2025). Tested on Artifactory 7.104.7 with Terraform 1.10.5 and OpenTofu 1.9.0

IMPROVEMENTS:

* resource/artifactory_local_helm_repository: Add attributes `force_non_duplicate_chart` and `force_metadata_name_version` to support new settings introduced in Artifactory v7.104.5. Issue: [#1199](https://github.com/jfrog/terraform-provider-artifactory/issues/1199) PR: [#1207](https://github.com/jfrog/terraform-provider-artifactory/pull/1207)

BUG FIXES:

* resource/artifactory_remote_docker_v2_repository: Fix incorrect default value for `block_pushing_schema1` and `tag_retention`. Issue: [#1186](https://github.com/jfrog/terraform-provider-artifactory/issues/1186) PR: [#1201](https://github.com/jfrog/terraform-provider-artifactory/pull/1201)
* resource/artifactory_\*\_repository: Improve state handling of `project_environments` attribute. Issue: [#1186](https://github.com/jfrog/terraform-provider-artifactory/issues/1186) PR: [#1201](https://github.com/jfrog/terraform-provider-artifactory/pull/1201)

## 12.8.4 (Februry 13, 2025). Tested on Artifactory 7.104.6 with Terraform 1.10.5 and OpenTofu 1.9.0

BUG FIXES:

* resource/artifactory_\*\_repository: Fix error when provider tries to refresh the resource for repository that is no longer existed on Artifactory. Issue: [#1189](https://github.com/jfrog/terraform-provider-artifactory/issues/1189) PR: [#1190](https://github.com/jfrog/terraform-provider-artifactory/pull/1190)
* resource/artifactory_local_debian_repository: Fix state drift when `index_compression_formats` attribute is not set in configuration. Added default value of `bz2`. Issue: [#1183](https://github.com/jfrog/terraform-provider-artifactory/issues/1183) PR: [#1195](https://github.com/jfrog/terraform-provider-artifactory/pull/1195)
* resource/artifactory_artifactory_release_bundle_custom_webhook, resource/artifactory_destination_custom_webhook, resource/artifactory_distribution_custom_webhook, resource/artifactory_release_bundle_custom_webhook: Fix provider panic crash. Issue: [#1192](https://github.com/jfrog/terraform-provider-artifactory/issues/1192) PR: [#1196](https://github.com/jfrog/terraform-provider-artifactory/pull/1196)

## 12.8.3 (January 28, 2025). Tested on Artifactory 7.98.14 with Terraform 1.10.5 and OpenTofu 1.9.0

NOTES:

If you are upgrading to this version from before 12.8.0 and encountered the following error message, please upgrade to 12.8.0 first. Then upgrade to the latest.

```
│ Error: Unable to Upgrade Resource State
│ 
│   with artifactory_remote_gems_repository.remote_gems_repo["DEV"],
│   on remote_repositories.tf line 360, in resource "artifactory_remote_gems_repository" "remote_gems_repo":
│  360: resource "artifactory_remote_gems_repository" "remote_gems_repo" {
│ 
│ This resource was implemented without an UpgradeState() method, however
│ Terraform was expecting an implementation for version 3 upgrade.
│ 
│ This is always an issue with the Terraform Provider and should be reported
│ to the provider developer.
╵
```

IMPROVEMENTS:

* resource/artifactory_remote_\*\_repository are migrated to Plugin Framework. PR: [#1180](https://github.com/jfrog/terraform-provider-artifactory/pull/1180)

## 12.8.2 (January 14, 2025). Tested on Artifactory 7.98.13 with Terraform 1.10.4 and OpenTofu 1.9.0

IMPROVEMENTS:

* resource/artifactory_local_alpine_repository, resource/artifactory_local_ansible_repository, resource/artifactory_local_cargo_repository, resource/artifactory_local_conan_repository, resource/artifactory_local_docker_v1_repository, resource/artifactory_local_docker_v2_repository, resource/artifactory_local_gradle_repository, resource/artifactory_local_helmoci_repository, resource/artifactory_local_ivy_repository, resource/artifactory_local_nuget_repository, resource/artifactory_local_helm_repository, resource/artifactory_local_hunggingfaceml_repository, resource/artifactory_local_oci_repository, resource/artifactory_local_rpm_repository, resource/artifactory_local_sbt_repository, resource/artifactory_local_terraform_module_repository, resource/artifactory_local_terraform_provider_repository are migrated to Plugin Framework. PR: [#1168](https://github.com/jfrog/terraform-provider-artifactory/pull/1168)

## 12.8.1 (January 13, 2025). Tested on Artifactory 7.98.13 with Terraform 1.10.4 and OpenTofu 1.9.0

BUG FIXES:

* resource/artifactory_artifact: Fix incorrect integer parsing for `size` attribute. Issue: [#1169](https://github.com/jfrog/terraform-provider-artifactory/issues/1169) PR: [#1171](https://github.com/jfrog/terraform-provider-artifactory/pull/1171)

## 12.8.0 (January 8, 2025). Tested on Artifactory 7.98.13 with Terraform 1.10.3 and OpenTofu 1.8.8

FEATURES:

**New Resource:** `artifactory_federated_huggingfaceml_repository` to support federated HuggingFace ML repository. PR: [#1161](https://github.com/jfrog/terraform-provider-artifactory/pull/1161)

IMPROVEMENTS:

* resource/artifactory_\*\_custom_webhook: Add attribute `method` to support setting HTTP method. PR: [#1160](https://github.com/jfrog/terraform-provider-artifactory/pull/1160) and [#1163](https://github.com/jfrog/terraform-provider-artifactory/pull/1163)
* resource/artifactory_general_security: Add attribute `encryption_policy` to support setting password encryption policy. Issue: [#1159](https://github.com/jfrog/terraform-provider-artifactory/issues/1159) PR: [#1162](https://github.com/jfrog/terraform-provider-artifactory/pull/1162)

BUG FIXES:

* resource/artifactory_artifact: Fix artifact upload incorrectly (using multipart form vs raw binary data). Issue: [#1083](https://github.com/jfrog/terraform-provider-artifactory/issues/1083) PR: [#1164](https://github.com/jfrog/terraform-provider-artifactory/pull/1164)

## 12.7.1 (December 18, 2024). Tested on Artifactory 7.98.11 with Terraform 1.10.2 and OpenTofu 1.8.7

BUG FIXES:

* resource/artifactory_remote_cargo_repository: Fix `git_registry_url` attribute to be incorrectly set to required. Issue: [#1153](https://github.com/jfrog/terraform-provider-artifactory/issues/1153) PR: [#1154](https://github.com/jfrog/terraform-provider-artifactory/pull/1154) and [#1155](https://github.com/jfrog/terraform-provider-artifactory/pull/1155)

## 12.7.0 (December 17, 2024). Tested on Artifactory 7.98.11 with Terraform 1.10.2 and OpenTofu 1.8.7

FEATURES:

**New Resource:** `artifactory_local_machinelearning_repository` to support local Machine Learning repository. PR: [#1152](https://github.com/jfrog/terraform-provider-artifactory/pull/1152)

IMPROVEMENTS:

* resource/artifactory_local_bower_repository, resource/artifactory_local_chef_repository, resource/artifactory_local_cocoapods_repository, resource/artifactory_local_composer_repository, resource/artifactory_local_conda_repository, resource/artifactory_local_cran_repository, resource/artifactory_local_gems_repository, resource/artifactory_local_generic_repository, resource/artifactory_local_gitlfs_repository, resource/artifactory_local_go_repository, resource/artifactory_local_helm_repository, resource/artifactory_local_hunggingfaceml_repository, resource/artifactory_local_npm_repository, resource/artifactory_local_opkg_repository, resource/artifactory_local_pup_repository, resource/artifactory_local_puppet_repository, resource/artifactory_local_pypi_repository, resource/artifactory_local_swift_repository, resource/artifactory_local_terraformbackend_repository, resource/artifactory_local_vagrant_repository are migrated to Plugin Framework. PR: [#1152](https://github.com/jfrog/terraform-provider-artifactory/pull/1152)

## 12.6.0 (December 13, 2024). Tested on Artifactory 7.98.10 with Terraform 1.10.2 and OpenTofu 1.8.7

FEATURES:

**New Resource:** `artifactory_archive_policy` to support upcoming Archive Policy feature. PR: [#1146](https://github.com/jfrog/terraform-provider-artifactory/pull/1146)

IMPROVEMENTS:

* resource/artifactory_artifact: Add attribute `content_base64` to support uploading data in base64 format string. Issue: [#1083](https://github.com/jfrog/terraform-provider-artifactory/issues/1083) PR: [#1149](https://github.com/jfrog/terraform-provider-artifactory/pull/1149)

BUG FIXES:

* resource/artifactory_backup: Add size validation to `excluded_repositories` attribute to ensure at least 1 item. Issue: [#1143](https://github.com/jfrog/terraform-provider-artifactory/issues/1143) PR: [#1147](https://github.com/jfrog/terraform-provider-artifactory/pull/1147)

## 12.5.1 (November 22, 2024)

BUG FIXES:

* dependency: Update Resty to 1.26.2. Potentially fix intermittent HTTP authentication issue. Issue: [#1135](https://github.com/jfrog/terraform-provider-artifactory/issues/1135) PR: [#1136](https://github.com/jfrog/terraform-provider-artifactory/pull/1136)

## 12.5.0 (November 15, 2024). Tested on Artifactory 7.98.8 with Terraform 1.9.8 and OpenTofu 1.8.5

NOTES:

* resource/artifactory_group: This resource is being deprecated and replaced by new [platform_group](https://registry.terraform.io/providers/jfrog/platform/latest/docs/resources/group) resource in the Platform provider. PR: [#1130](https://github.com/jfrog/terraform-provider-artifactory/pull/1130)

## 12.4.1 (November 11, 2024). Tested on Artifactory 7.98.8 with Terraform 1.9.8 and OpenTofu 1.8.5

BUG FIXES:

* resource/artifactory_build_webhook, resource/artifactory_release_bundle_webhook, resource/artifactory_release_bundle_v2_webhook, resource/artifactory_artifact_webhook, resource/artifactory_artifact_property_webhook, resource/artifactory_docker_webhook, resource/artifactory_build_custom_webhook, resource/artifactory_release_bundle_custom_webhook, resource/artifactory_release_bundle_v2_custom_webhook, resource/artifactory_artifact_custom_webhook, resource/artifactory_artifact_property_custom_webhook, resource/artifactory_docker_custom_webhook, resource/artifactory_ldap_group_setting_v2, resource/artifactory_repository_layout, resource/artifactory_release_bundle_v2, resource/artifactory_vault_configuration, resource/artifactory_user, resource/artifactory_managed_user, resource/artifactory_unmanaged_user: Fix attribute validation not working for unknown value (e.g. when resource is used in a module). Issue: [#1120](https://github.com/jfrog/terraform-provider-artifactory/issues/1120) PR: [#1123](https://github.com/jfrog/terraform-provider-artifactory/pull/1123)
* resource/artifactory_package_cleanup_policy: Relax validations to allow 0 for `created_before_in_months` and `last_downloaded_before_in_months` attributes. Add validation run to ensure both can't be 0 at the same time. Issue: [#1122](https://github.com/jfrog/terraform-provider-artifactory/issues/1122) PR: [#1126](https://github.com/jfrog/terraform-provider-artifactory/pull/1126)

## 12.4.0 (November 4, 2024). Tested on Artifactory 7.98.7 with Terraform 1.9.8 and OpenTofu 1.8.5

IMPROVEMENTS:

* resource/artifactory_remote_go_repository: Update list of valid values for attribute `vcs_git_provider`. PR: [#1115](https://github.com/jfrog/terraform-provider-artifactory/pull/1115)

## 12.3.3 (November 1, 2024). Tested on Artifactory 7.98.7 with Terraform 1.9.8 and OpenTofu 1.8.4

BUG FIXES:

* resource/artifactory_local_repository_single_replication: Fix unknown value error for `check_binary_existence_in_filestore`. Issue: [#1105](https://github.com/jfrog/terraform-provider-artifactory/issues/1105) PR: [#1113](https://github.com/jfrog/terraform-provider-artifactory/pull/1113)

## 12.3.2 (October 30, 2024). Tested on Artifactory 7.98.7 with Terraform 1.9.8 and OpenTofu 1.8.4

IMPROVEMENTS:

* resource/artifactory_package_cleanup_policy: Update valid values for `package_types` attribute. PR: [#1107](https://github.com/jfrog/terraform-provider-artifactory/pull/1107)

BUG FIXES:

* resource/artifactory_mail_server: Fix error when unsetting an optional attribute. Issue: [#1103](https://github.com/jfrog/terraform-provider-artifactory/issues/1103) PR: [#1106](https://github.com/jfrog/terraform-provider-artifactory/pull/1106)
* resource/artifactory_remote_repository_replication: Fix unknown value error for `check_binary_existence_in_filestore`. Issue: [#1105](https://github.com/jfrog/terraform-provider-artifactory/issues/1105) PR: [#1107](https://github.com/jfrog/terraform-provider-artifactory/pull/1107)

## 12.3.1 (October 18, 2024). Tested on Artifactory 7.90.14 with Terraform 1.9.8 and OpenTofu 1.8.3

BUG FIXES:

* resource/artifactory_local_repository_multi_replication: Fix error when updating resource with new `cron_exp` value. Issue: [#1099](https://github.com/jfrog/terraform-provider-artifactory/issues/1099) PR: [#1100](https://github.com/jfrog/terraform-provider-artifactory/pull/1100)

## 12.3.0 (October 16, 2024). Tested on Artifactory 7.90.14 with Terraform 1.9.7 and OpenTofu 1.8.3

IMPROVEMENTS:

* provider: Add `tfc_credential_tag_name` configuration attribute to support use of different/[multiple Workload Identity Token in Terraform Cloud Platform](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/manual-generation#generating-multiple-tokens). Issue: [#68](https://github.com/jfrog/terraform-provider-shared/issues/68) PR: [#1097](https://github.com/jfrog/terraform-provider-artifactory/pull/1097)

## 12.2.0 (October 14, 2024). Tested on Artifactory 7.90.14 with Terraform 1.9.7 and OpenTofu 1.8.3

IMPROVEMENTS:

* resource/artifactory_local_repository_multi_replication: Add `disable_proxy` attribute to `replication` to support not using proxy. Issue: [#1088](https://github.com/jfrog/terraform-provider-artifactory/issues/1088) PR: [#1095](https://github.com/jfrog/terraform-provider-artifactory/pull/1095)
* resource/artifactory_remote_conan_repository, resource/artifactory_remote_gems_repository, resource/artifactory_remote_go_repository, resource/artifactory_remote_gradle_repository, resource/artifactory_remote_huggingfaceml_repository, resource/artifactory_remote_nuget_repository: Add `curated` attribute to support enabled Curation for these package types. Issue: [#1091](https://github.com/jfrog/terraform-provider-artifactory/issues/1091) PR: [#1096](https://github.com/jfrog/terraform-provider-artifactory/pull/1096)

## 12.1.1 (October 2, 2024). Tested on Artifactory 7.90.13 with Terraform 1.9.6 and OpenTofu 1.8.2

IMPROVEMENTS:

* resource/artifactory_\*\_webhook is migrated to Plugin Framework. PR: [#1087](https://github.com/jfrog/terraform-provider-artifactory/pull/1087)
* resource/artifactory_\*\_custom_webhook is migrated to Plugin Framework. PR: [#1089](https://github.com/jfrog/terraform-provider-artifactory/pull/1089)

## 12.1.0 (September 26, 2024). Tested on Artifactory 7.90.10 with Terraform 1.9.6 and OpenTofu 1.8.2

IMPROVEMENTS:

* **New Resource:** artifactory_virtual_cocoapods_repository, which was missing from the provider. Issue: [#1079](https://github.com/jfrog/terraform-provider-artifactory/issues/1079) PR: [#1084](https://github.com/jfrog/terraform-provider-artifactory/pull/1084)
* resource/artifactory_remote_generic_repository: Add `retrieve_sha256_from_server` attribute. Issue: [#1080](https://github.com/jfrog/terraform-provider-artifactory/issues/1080) PR: [#1085](https://github.com/jfrog/terraform-provider-artifactory/pull/1085)

## 12.0.0 (September 16, 2024). Tested on Artifactory 7.90.10 with Terraform 1.9.6 and OpenTofu 1.8.2

BREAKING CHANGES:

* resource/artifactory_remote_\*\_repository: Fix mismatch between documentation and actual default vaule for `list_remote_folder_items` attribute. This will also match the change (default to `false`) introduced to the REST API in Artifactory v7.94.0. **Note:** Due to limitation of the Terraform SDKv2 framework, it is not possible to detect the attribute was not set and the state has the (previous) default value of `true`, then automaticaly upgrades it to `false`. Therefore this update will introduce state drift if this attribute is not set in your configuration. Use `terraform apply -refresh-only` to update your Terraform states to match. PR: [#1072](https://github.com/jfrog/terraform-provider-artifactory/pull/1072)
* provider: Removed deprecated `check_license` attribute. PR: [#1073](https://github.com/jfrog/terraform-provider-artifactory/pull/1073)
* resource/artifactory_access_token: Removed deprecated resource. PR: [#1073](https://github.com/jfrog/terraform-provider-artifactory/pull/1073)
* resource/artifactory_replication_config: Removed deprecated resource. PR: [#1073](https://github.com/jfrog/terraform-provider-artifactory/pull/1073)
* resource/artifactory_single_replication_config: Removed deprecated resource. PR: [#1073](https://github.com/jfrog/terraform-provider-artifactory/pull/1073)

## 11.9.2 (September 12, 2024). Tested on Artifactory 7.90.10 with Terraform 1.9.5 and OpenTofu 1.8.2

IMPROVEMENTS:

* provider: Upgrade Golang version to 1.22.7 due to CVE-2024-34156. Security Advisory: [GHSA-wcq6-89h8-g366](https://github.com/jfrog/terraform-provider-artifactory/security/advisories/GHSA-wcq6-89h8-g366)
* resource/artifactory_package_cleanup_policy: Remove version limitation note from documentation.

PR: [#1071](https://github.com/jfrog/terraform-provider-artifactory/pull/1071)

## 11.9.1 (September 9, 2024). Tested on Artifactory 7.90.9 with Terraform 1.9.5 and OpenTofu 1.8.2

IMPROVEMENTS:

* resource/artifactory_package_cleanup_policy: Add beta warning message to documentation. PR: [#1068](https://github.com/jfrog/terraform-provider-artifactory/pull/1068)

## 11.9.0 (September 6, 2024). Tested on Artifactory 7.90.9 with Terraform 1.9.5 and OpenTofu 1.8.2

IMPROVEMENTS:

* resource/artifactory_package_cleanup_policy: Add `project_key` attribute. Update attribute validations and documentation. PR: [#1065](https://github.com/jfrog/terraform-provider-artifactory/pull/1065)

## 11.8.0 (September 3, 2024). Tested on Artifactory 7.90.9 with Terraform 1.9.5 and OpenTofu 1.8.1

IMPROVEMENTS:

* resource/artifactory_federated_\*\_repository: Add `access_token` attribute to `member` to support `cleanup_on_delete` for JPD setup without Access Federation for access token. Also improve error handling when deleting member federated repository. PR: [#1057](https://github.com/jfrog/terraform-provider-artifactory/pull/1057)
* resource/artifactory_local_repository_single_replication is migrated to Plugin Framework. PR: [#1059](https://github.com/jfrog/terraform-provider-artifactory/pull/1059)
* resource/artifactory_remote_repository_single_replication is migrated to Plugin Framework. PR: [#1060](https://github.com/jfrog/terraform-provider-artifactory/pull/1060)
* resource/artifactory_local_repository_multi_replication is migrated to Plugin Framework. PR: [#1061](https://github.com/jfrog/terraform-provider-artifactory/pull/1061)

## 11.7.0 (August 22, 2024). Tested on Artifactory 7.90.8 with Terraform 1.9.5 and OpenTofu 1.8.1

FEATURES:

**New Resource:** `artifactory_package_cleanup_policy` to support [Retention Policy](https://jfrog.com/help/r/jfrog-platform-administration-documentation/retention-policies) PR: [#1056](https://github.com/jfrog/terraform-provider-artifactory/pull/1056)

## 11.6.0 (August 12, 2024). Tested on Artifactory 7.90.7 with Terraform 1.9.4 and OpenTofu 1.8.1

NOTES:

* resource/artifactory_saml_settings: This resource is being deprecated and replaced by new [platform_saml_settings](https://registry.terraform.io/providers/jfrog/platform/latest/docs/resources/saml_settings) resource in the Platform provider. PR: [#1052](https://github.com/jfrog/terraform-provider-artifactory/pull/1052)

BUG FIXES:

* resource/artifactory_virtual_go_repository: Fix updating `external_dependencies_patterns` attribute triggers a resource replacement that wasn't fixed in PR [#1005](https://github.com/jfrog/terraform-provider-artifactory/pull/1005). Issue: [#1004](https://github.com/jfrog/terraform-provider-artifactory/issues/1004) PR: [#1051](https://github.com/jfrog/terraform-provider-artifactory/pull/1051)

## 11.5.1 (August 8, 2024). Tested on Artifactory 7.90.6 with Terraform 1.9.4 and OpenTofu 1.8.1

BUG FIXES:

* resource/artifactory_item_properties: Fix destroying resource did not remove properties in Artfiactory. Issue: [#1049](https://github.com/jfrog/terraform-provider-artifactory/issues/1049) PR: [#1050](https://github.com/jfrog/terraform-provider-artifactory/pull/1050)

## 11.5.0 (August 6, 2024). Tested on Artifactory 7.90.6 with Terraform 1.9.3 and OpenTofu 1.8.0

FEATURES:

**New Resource:** `artifactory_item_properties` Issue: [#1041](https://github.com/jfrog/terraform-provider-artifactory/issues/1041) PR: [#1046](https://github.com/jfrog/terraform-provider-artifactory/pull/1046)

IMPROVEMENTS:

* resource/artifactory_remote_gradle_repository, resource/artifactory_remote_ivy_repository, resource/artifactory_remote_maven_repository, resource/artifactory_remote_sbt_repository: Add `max_unique_snapshots` attribute support. Issue: [#1039](https://github.com/jfrog/terraform-provider-artifactory/issues/1039) PR: [#1043](https://github.com/jfrog/terraform-provider-artifactory/pull/1043)

## 11.4.0 (July 31, 2024). Tested on Artifactory 7.90.5 with Terraform 1.9.3 and OpenTofu 1.8.0

FEATURES:

**New Resource:**

* `artifactory_release_bundle_v2`
* `artifactory_release_bundle_v2_promotion`

PR: [#1040](https://github.com/jfrog/terraform-provider-artifactory/pull/1040)

## 11.3.0 (July 26, 2024). Tested on Artifactory 7.90.5 with Terraform 1.9.3 and OpenTofu 1.7.3

FEATURES:

**New Data Source and Resource:**

* `artifactory_local_ansible_repository`
* `artifactory_remote_ansible_repository`
* `artifactory_virtual_ansible_repository`
* `artifactory_federated_ansible_repository`

`Ansible` package type is supported in Artfactory 7.90.1 and later. 

## 11.2.2 (July 25, 2024). Tested on Artifactory 7.90.5 with Terraform 1.9.3 and OpenTofu 1.7.3

BUG FIXES:

* resource/artifactory_group: Fix updating `name` attribute results in API error. Updating this attribute now will trigger a deletion and recreation of the resource. Issue: [#1035](https://github.com/jfrog/terraform-provider-artifactory/issues/1035) PR: [#1037](https://github.com/jfrog/terraform-provider-artifactory/pull/1037)
* resource/artifactory_scoped_token: Fix inconsistent value for `ignore_missing_token_warning` attribute when it is set. Issue: [#1034](https://github.com/jfrog/terraform-provider-artifactory/issues/1034) PR: [#1038](https://github.com/jfrog/terraform-provider-artifactory/pull/1038)

## 11.2.1 (July 22, 2024). Tested on Artifactory 7.84.17 with Terraform 1.9.2 and OpenTofu 1.7.3

BUG FIXES:

* resource/artifactory_user, resource/artifactory_managed_user, resource/artifactory_unmanaged_user: Fix `password` validation for interpolated value. Also improve validation logic. Issue: [#1031](https://github.com/jfrog/terraform-provider-artifactory/issues/1031) PR: [#1032](https://github.com/jfrog/terraform-provider-artifactory/pull/1032)

## 11.2.0 (July 16, 2024). Tested on Artifactory 7.84.17 with Terraform 1.9.2 and OpenTofu 1.7.3

FEATURES:

* **New Resource:**
  * `artifactory_destination_webhook`
  * `artifactory_user_webhook`
  * `artifactory_release_bundle_v2_webhook`
  * `artifactory_release_bundle_v2_promotion_webhook`
  * `artifactory_artifact_lifecycle_webhook`
  * `artifactory_destination_custom_webhook`
  * `artifactory_user_custom_webhook`
  * `artifactory_release_bundle_v2_custom_webhook`
  * `artifactory_release_bundle_v2_promotion_custom_webhook`
  * `artifactory_artifact_lifecycle_custom_webhook`

  Issue: [#1012](https://github.com/jfrog/terraform-provider-artifactory/issues/1012) 
  PR: [#1019](https://github.com/jfrog/terraform-provider-artifactory/pull/1019)
* resource/artifactory_oauth_settings, resource/artifactory_saml_settings: Remove warning message about undocumented APIs. Issue: [#291](https://github.com/jfrog/terraform-provider-artifactory/issues/291) PR: [#1020](https://github.com/jfrog/terraform-provider-artifactory/pull/1020)
* resource/artifactory_user, resource/artifactory_managed_user: Add `password_policy` attribute to support configurable password validation. Issue: [#959](https://github.com/jfrog/terraform-provider-artifactory/issues/959) PR: [#1024](https://github.com/jfrog/terraform-provider-artifactory/pull/1024)
* resource/artifactory_scoped_token: Add attribute `ignore_missing_token_warning` to hide warning message about missing token when refreshing state from Artifactory. Issue: [#1021](https://github.com/jfrog/terraform-provider-artifactory/issues/1021) PR: [#1026](https://github.com/jfrog/terraform-provider-artifactory/pull/1026)

BUG FIXES:

* data/artifactory_\*\_repository: Fix 400 error if repository does not exist. Issue: [#1018](https://github.com/jfrog/terraform-provider-artifactory/issues/1018) PR: [#1022](https://github.com/jfrog/terraform-provider-artifactory/pull/1022)

NOTES:

* resource/artifactory_artifactory_release_bundle_webhook: This resource is being deprecated and replaced by new `artifactory_destination_webhook` resource
* resource/artifactory_artifactory_release_bundle_custom_webhook: This resource is being deprecated and replaced by new `artifactory_destination_custom_webhook` resource

## 11.1.0 (June 26, 2024). Tested on Artifactory 7.84.15 with Terraform 1.8.5 and OpenTofu 1.7.2

NOTES:

* provider: `check_license` attribute is deprecated and provider no longer checks Artifactory license during initialization. It will be removed in the next major version release.

FEATURES:

* **New Resource:** `artifactory_vault_configuration` PR: [#1008](https://github.com/jfrog/terraform-provider-artifactory/pull/1008)
* resource/artifactory_remote_\*\_repository: Add `archive_browsing_enabled` attribute to all remote repository resources. Issue: [#999](https://github.com/jfrog/terraform-provider-artifactory/issues/999) PR: [#1003](https://github.com/jfrog/terraform-provider-artifactory/pull/1003)

BUG FIXES:

* resource/artifactory_virtual_bower_repository, resource/artifactory_virtual_npm_repository: Fix `external_dependencies_patterns` attribute set to be force new, which causes the resource to be recreated when attribute value changes. Issue: [#1004](https://github.com/jfrog/terraform-provider-artifactory/issues/1004) PR: [#1005](https://github.com/jfrog/terraform-provider-artifactory/pull/1005)
* resource/artifactory_\*\_user: Allow `+` character for `name` attribute. This is allowed in SaaS instance but currently not for self-hosted. PR: [#1007](https://github.com/jfrog/terraform-provider-artifactory/pull/1007)

## 11.0.0 (June 6, 2024). Tested on Artifactory 7.84.14 with Terraform 1.8.5 and OpenTofu 1.7.2

BREAKING CHANGES:

* Resources `artifactory_access_token`, `artifactory_replication_config`, and `artifactory_single_replication_config` have been removed from provider. PR: [#995](https://github.com/jfrog/terraform-provider-artifactory/pull/995)

## 10.8.4 (June 5, 2024)

BUG FIXES:

* resource/artifactory_build_webhook, resource/artifactory_custom_build_webhook: Fix criteria validation to allow `include_patterns` attribute values with `any_build` attribute set to `false`. Issue: [#987](https://github.com/jfrog/terraform-provider-artifactory/issues/987) PR: [#993](https://github.com/jfrog/terraform-provider-artifactory/pull/993)

## 10.8.3 (June 3, 2024)

BUG FIXES:

* resource/artifactory_scoped_token: Fix incorrect validation with actions values for `scopes` attribute. Issue: [#985](https://github.com/jfrog/terraform-provider-artifactory/issues/985) PR: [#986](https://github.com/jfrog/terraform-provider-artifactory/pull/986)

IMPROVEMENTS:

* Documentation: Move `metadata_retrieval_timeout_secs` attribute documentation from `artifactory_remote_maven_repository` to "Artifactory Remote Repository Common Arguments" documentation. Issue: [#983](https://github.com/jfrog/terraform-provider-artifactory/issues/983) PR: [#984](https://github.com/jfrog/terraform-provider-artifactory/pull/984)

## 10.8.2 (May 31, 2024)

BUG FIXES:

* resource/artifactory_keypair: Remove `private_key` value from warning and error messages. Issue: [#977](https://github.com/jfrog/terraform-provider-artifactory/issues/977) PR: [#979](https://github.com/jfrog/terraform-provider-artifactory/pull/979)
* resource/artifactory_scoped_token: Add check for status code 404 after resource creation and display warning message due to Persistency Threshold. Issue: [#980](https://github.com/jfrog/terraform-provider-artifactory/issues/980) PR: [#981](https://github.com/jfrog/terraform-provider-artifactory/pull/981)

## 10.8.1 (May 24, 2024)

BUG FIXES:

* resource/artifactory_\*\_webhook, resource/artifactory_\*\_custom_webhook: Fix various crashes when importing the resource with optional attributes not set in the configuration. PR: [#973](https://github.com/jfrog/terraform-provider-artifactory/pull/973)

## 10.8.0 (May 20, 2024)

IMPROVEMENTS:

* resource/artifactory_scoped_token: Add support for project admin token scope introduced in Artifactory 7.84.3. See [Access Token Creation by Project Admins](https://jfrog.com/help/r/jfrog-platform-administration-documentation/access-token-creation-by-project-admins) for more details. PR: [#965](https://github.com/jfrog/terraform-provider-artifactory/pull/965)

## 10.7.7 (May 17, 2024). Tested on Artifactory 7.84.10 with Terraform CLI v1.8.3

BUG FIXES:

* provider: Fix inability to use `api_key` attribute without also setting `access_token` attribute. Issue: [#966](https://github.com/jfrog/terraform-provider-artifactory/issues/966) PR: [#967](https://github.com/jfrog/terraform-provider-artifactory/pull/967)

## 10.7.6 (May 10, 2024)

BUG FIXES:

* resource/artifactory_managed_user: Update `password` minimum length validation to 8 characters which matches default length for both cloud and self-hosted versions. Issue: [#959](https://github.com/jfrog/terraform-provider-artifactory/issues/959) PR: [#962](https://github.com/jfrog/terraform-provider-artifactory/pull/962)

## 10.7.5 (May 2, 2024)

IMPROVEMENTS:

* resource/artifactory_general_security is migrated to Plugin Framework. PR: [#948](https://github.com/jfrog/terraform-provider-artifactory/pull/948)

## 10.7.4 (May 1, 2024)

IMPROVEMENTS:

* resource/artifactory_virtual_npm_repository: Add documentation for missing attributes `external_dependencies_enabled`, `external_dependencies_remote_repo`, and `external_dependencies_patterns` PR: [#947](https://github.com/jfrog/terraform-provider-artifactory/pull/947)

## 10.7.3 (Apr 31, 2024)

BUG FIXES:

* resource/artifactory_managed_user, resource/artifactory_unmanaged_user, resource/artifactory_user: Make `name` attribute trigger resource replacement if changed. Issue: [#944](https://github.com/jfrog/terraform-provider-artifactory/issues/944) PR: [#946](https://github.com/jfrog/terraform-provider-artifactory/pull/946)

## 10.7.2 (Apr 26, 2024)

BUG FIXES:

* resource/artifactory_proxy: Fix hidden state drifts with resource created using SDKv2 (i.e. <= 10.1.0). Issue: [#941](https://github.com/jfrog/terraform-provider-artifactory/issues/941) PR: [#943](https://github.com/jfrog/terraform-provider-artifactory/pull/943)

## 10.7.1 (Apr 25, 2024)

BUG FIXES:

* resource/artifactory_managed_user, resource/artifactory_unmanaged_user, resource/artifactory_user: Toggle between using (old) Artifactory Security API and (new) Access API based on Artifactory version 7.84.3 due to Access API bug in updating user without password field. Issue: [#931](https://github.com/jfrog/terraform-provider-artifactory/issues/931) PR: [#940](https://github.com/jfrog/terraform-provider-artifactory/pull/940)

## 10.7.0 (Apr 18, 2024)

FEATURES:

* provider: Add support for Terraform Cloud Workload Identity Token. PR: [#938](https://github.com/jfrog/terraform-provider-artifactory/pull/938)

## 10.6.2 (Apr 17, 2024)

BUG FIXES:

* resource/artifactory_unmanaged_user, resource/artifactory_user: Revert storing auto-generated `password` attribute value in Terraform state. Revert back to Artifactory Security API until Artifactory version 7.83.1 due to Access API bug in updating user without password field. Issue: [#931](https://github.com/jfrog/terraform-provider-artifactory/issues/931) PR: [#937](https://github.com/jfrog/terraform-provider-artifactory/pull/937)

IMPROVEMENTS:

* resource/artifactory_property_set is migrated to Plugin Framework. PR: [#935](https://github.com/jfrog/terraform-provider-artifactory/pull/935)
* resource/artifactory_repository_layout is migrated to Plugin Framework. PR: [#936](https://github.com/jfrog/terraform-provider-artifactory/pull/936)

## 10.6.1 (Apr 12, 2024)

BUG FIXES:

* provider: Fix `check_license` attribute not process correctly. Issue: [#930](https://github.com/jfrog/terraform-provider-artifactory/issues/930) PR: [#934](https://github.com/jfrog/terraform-provider-artifactory/pull/934)

## 10.6.0 (Apr 12, 2024)

FEATURES:

* **New Resource:** `artifactory_artifact` for uploading artifact to repository. Issue: [#896](https://github.com/jfrog/terraform-provider-artifactory/issues/896) PR: [#933](https://github.com/jfrog/terraform-provider-artifactory/pull/933)

BUG FIXES:

* provider: Fix crash when provider checks for Artifactory license fails due to networking issue. Issue: [#930](https://github.com/jfrog/terraform-provider-artifactory/issues/930) PR: [#933](https://github.com/jfrog/terraform-provider-artifactory/pull/933)

## 10.5.1 (Apr 10, 2024)

BUG FIXES:

* resource/artifactory_user, resource/artifactory_managed_user: Fix error when updating resource without `password` attribute defined (i.e. relying on auto generated password). Issue: [#931](https://github.com/jfrog/terraform-provider-artifactory/issues/931) PR: [#932](https://github.com/jfrog/terraform-provider-artifactory/pull/932)

## 10.5.0 (Apr 9, 2024)

FEATURES:

* **New Resource:** `artifactory_password_expiration_policy` and `artifactory_user_lock_policy`

Issue: [#927](https://github.com/jfrog/terraform-provider-artifactory/issues/927) PR: [#928](https://github.com/jfrog/terraform-provider-artifactory/pull/928)

## 10.4.4 (Apr 8, 2024)

BUG FIXES:

* resource/artifactory_virtual_helmoci_repository: Fix incorrect package type. Issue: [#925](https://github.com/jfrog/terraform-provider-artifactory/issues/925) PR: [#926](https://github.com/jfrog/terraform-provider-artifactory/pull/926)

## 10.4.3 (Apr 1, 2024)

BUG FIXES:

* resource/artifactory_user, resource/artifactory_managed_user, resource/artifactory_unmanaged_user: Restore backward compatibility with Artifactory 7.49.2 and earlier. Issue: [#922](https://github.com/jfrog/terraform-provider-artifactory/issues/922) PR: [#923](https://github.com/jfrog/terraform-provider-artifactory/pull/923)

## 10.4.2 (Mar 27, 2024)

BUG FIXES:

* Provider: Further improvement on API error handling that was missed in previous updates. Issue: [#886](https://github.com/jfrog/terraform-provider-artifactory/issues/885) PR: [#921](https://github.com/jfrog/terraform-provider-artifactory/pull/921)

## 10.4.1 (Mar 25, 2024)

BUG FIXES:

* resource/artifactory_user, resource/artifactory_managed_user, resource/artifactory_unmanaged_user: Fix `groups` handling to keep groups membership from Artifactory sychronized and avoid state drift. Also fix inability to create a user without `password` set with `internal_password_disabled` is set to `true`. Issue: [#915](https://github.com/jfrog/terraform-provider-artifactory/issues/915) PR: [#920](https://github.com/jfrog/terraform-provider-artifactory/pull/920)

## 10.4.0 (Mar 20, 2024)

IMPROVEMENTS:

* resource/artifactory_\*\_webhook and resource/artifactory_\*\_custom_webhook: Add `any_federated` attribute for `artifact`, `artifact_property`, and `docker` critera. Issue: [#906](https://github.com/jfrog/terraform-provider-artifactory/issues/906) PR: [#916](https://github.com/jfrog/terraform-provider-artifactory/pull/916)
* resource/artifactory_*_repository: Update `key` validation to allow number at the beginning. Also apply length validation to match Artifactory UI. PR: [#917](https://github.com/jfrog/terraform-provider-artifactory/pull/917)

## 10.3.3 (Mar 18, 2024)

NOTES:

* This patch release has no functional changes. This release switches the Terraform registry signing key to the same key pair used by the other JFrog providers. This enable all providers to be signed and installable on OpenTofu registry which allows only one signing key (Terraform allows multiples).

Issue: [#909](https://github.com/jfrog/terraform-provider-artifactory/issues/909) PR: [#910](https://github.com/jfrog/terraform-provider-artifactory/pull/910)

## 10.3.2 (Mar 15, 2024)

BUG FIXES:

* resource/artifactory_user, resource/artifactory_managed_user, resource/artifactory_unmanaged_user: Fix `groups` handling so `readers` group from Artifactory no longer cause state drift. Issue: [#900](https://github.com/jfrog/terraform-provider-artifactory/issues/900)

PR: [#908](https://github.com/jfrog/terraform-provider-artifactory/pull/908)

## 10.3.1 (Mar 14, 2024)

BUG FIXES:

* resource/artifactory_scoped_token: Fix `scopes` attribute handling for scope with space character wraps in double quote (e.g. group name). Issue: [#903](https://github.com/jfrog/terraform-provider-artifactory/issues/903)
* resource/artifactory_api_key: Add API key deprecation notice to documentation (deprecation message already displays when uses via Terraform CLI).

PR: [#905](https://github.com/jfrog/terraform-provider-artifactory/pull/905)

NOTES:

* resource/artifactory_permission_target: Added deprecation notice. The new [`platform_permission` resource](https://registry.terraform.io/providers/jfrog/platform/latest/docs/resources/permission) in the JFrog Platform provider replace this resource. 

## 10.3.0 (Mar 11, 2024)

FEATURES:

* resource/artifactory_*_webhook: add `use_secret_for_signing` attribute for all non-custom webhook types. PR: [#902](https://github.com/jfrog/terraform-provider-artifactory/pull/902)

## 10.2.0 (Mar 6, 2024)

FEATURES:

* data/artifactory_*_oci_repository: add OCI package support for all repository type.
* resource/artifactory_*_oci_repository: add OCI package support for all repository type.

PR: [#897](https://github.com/jfrog/terraform-provider-artifactory/pull/897) Issue: [#885](https://github.com/jfrog/terraform-provider-artifactory/issues/885)

* data/artifactory_*_helmoci_repository: add Helm OCI package support for all repository type.
* resource/artifactory_*_helmoci_repository: add Helm OCI package support for all repository type.

PR: [#898](https://github.com/jfrog/terraform-provider-artifactory/pull/898) Issue: [#880](https://github.com/jfrog/terraform-provider-artifactory/issues/880)

## 10.1.5 (Feb 29, 2024)

BUG FIXES:

* resource/artifactory_scoped_token: Fix validation regex for `scopes` attribute. Also fix documentation with invalid Markdown. PR: [#895](https://github.com/jfrog/terraform-provider-artifactory/pull/895) Issue: [#889](https://github.com/jfrog/terraform-provider-artifactory/issues/889)

## 10.1.4 (Feb 14, 2024)

BUG FIXES:

* data/artifactory_virtual_maven_repository: Restore data source after being removed from the provider by mistake. PR: [#887](https://github.com/jfrog/terraform-provider-artifactory/pull/887) Issue: [#873](https://github.com/jfrog/terraform-provider-artifactory/issues/873)
* resource/artifactory_permission_target: Add check for '409 Conflict' error during resource creation and ignores it. PR: [#888](https://github.com/jfrog/terraform-provider-artifactory/pull/888) Issue: [#853](https://github.com/jfrog/terraform-provider-artifactory/issues/853)

## 10.1.3 (Feb 7, 2024)

BUG FIXES:

* resource/artifactory_group: Add length validation to `name` attribute to match Artifactory web UI. PR: [#884](https://github.com/jfrog/terraform-provider-artifactory/pull/884) Issue: [#883](https://github.com/jfrog/terraform-provider-artifactory/issues/883)

## 10.1.2 (Jan 19, 2024)

BUG FIXES:

* resource/artifactory_scoped_token: Fix `scopes` attribute validation for `actions` to include `x` and `s` values. PR: [#875](https://github.com/jfrog/terraform-provider-artifactory/pull/875)
* resource/artifactory_ldap_settings_v2: Fix validations for `search_filter`, `search_base`, `user_dn_pattern`, and `manager_dn` attributes when used with variables. PR: [#876](https://github.com/jfrog/terraform-provider-artifactory/pull/876) Issue: [#870](https://github.com/jfrog/terraform-provider-artifactory/issues/870)
* resource/artifactory_group_settings_v2: Fix validations for `group_base_dn` and `filter` attributes when used with variables. PR: [#876](https://github.com/jfrog/terraform-provider-artifactory/pull/876)

## 10.1.1 (Jan 17, 2024)

IMPROVEMENTS:

* resource/artifactory_proxy is migrated to Plugin Framework. PR: [#871](https://github.com/jfrog/terraform-provider-artifactory/pull/871)

## 10.1.0 (Jan 12, 2024)

FEATURES:

* resource/artifactory_remote_docker_repository: add new attribute `project_id`. PR: [#869](https://github.com/jfrog/terraform-provider-artifactory/pull/869)

IMPROVEMENTS:

* data source/artifactory_file: improve description for `path_is_aliased` attribute. PR: [#868](https://github.com/jfrog/terraform-provider-artifactory/pull/868)

## 10.0.2 (Dec 18, 2023). Tested on Artifactory 7.71.8 with Terraform CLI v1.6.6

IMPROVEMENTS:

* provider: downgrade Resty to 2.9.1 due to CVE in 2.10.0. PR: [#859](https://github.com/jfrog/terraform-provider-artifactory/pull/859)
* testing: improve resource drift error reporting.

## 10.0.1 (Dec 13, 2023). Tested on Artifactory 7.71.8 with Terraform CLI v1.6.5

IMPROVEMENTS:

* provider: remove Terraform protocol checks. PR: [#857](https://github.com/jfrog/terraform-provider-artifactory/pull/857)

## 10.0.0 (Dec 12, 2023). Tested on Artifactory 7.71.8 with Terraform CLI v1.6.5

BREAKING CHANGES:

* Terraform protocol version 5 is no longer supported. This provider will only work with Terraform protocol version 6 to take advantage of new functionality. This means only Terraform CLI 1.0 and above is supported.

FEATURES:

* data source/artifactory_repositories: add a new data source for retrieving list of repository, optionally filtered by repository type, package type, and project key. PR: [#839](https://github.com/jfrog/terraform-provider-artifactory/pull/839) Issue: [#716](https://github.com/jfrog/terraform-provider-artifactory/issues/716)
* data source/artifactory_file_list: add new data source to retrieve a list of artifacts from a repository. PR: [#855](https://github.com/jfrog/terraform-provider-artifactory/pull/855)

## 9.9.2 (Dec 5, 2023). Tested on Artifactory 7.71.5 with Terraform CLI v1.6.5

IMPROVEMENTS:

* Add telemetry to resources that were migrated to Plugin Framework PR: [#852](https://github.com/jfrog/terraform-provider-artifactory/pull/852)

## 9.9.1 (Dec 4, 2023). Tested on Artifactory 7.71.5 with Terraform CLI v1.6.5

BUG FIXES:

* Fix incorrect use of empty error message when API fails PR: [#850](https://github.com/jfrog/terraform-provider-artifactory/pull/850) Issue: [#849](https://github.com/jfrog/terraform-provider-artifactory/issues/849)
* resource/artifactory_global_environment: Fix incorrect environment from Artifactroy being matched and triggers a state drift. PR: [#851](https://github.com/jfrog/terraform-provider-artifactory/pull/851)

## 9.9.0 (Nov 17, 2023). Tested on Artifactory 7.71.4 with Terraform CLI v1.6.4

IMPROVEMENTS:

* resource/artifactory_federated_*_repository: Add `proxy` and `disable_proxy` attributes. PR: [#848](https://github.com/jfrog/terraform-provider-artifactory/pull/848) Issue: [#838](https://github.com/jfrog/terraform-provider-artifactory/issues/838)

## 9.8.0 (Nov 8, 2023). Tested on Artifactory 7.71.4 with Terraform CLI v1.6.3

IMPROVEMENTS:

* resource/artifactory_remote_docker_repository, resource/artifactory_remote_maven_repository, resource/artifactory_npm_docker_repository,resource/artifactory_remote_pypi_repository: Add `curated` attribute to support enabling the repository resource for Curation Service. Issue: [#831](https://github.com/jfrog/terraform-provider-artifactory/issues/831) PR: [#844](https://github.com/jfrog/terraform-provider-artifactory/pull/844)
* resource/artifactory_pull_replication: Add missing deprecation message to documentation. PR: [#843](https://github.com/jfrog/terraform-provider-artifactory/pull/843)

## 9.7.4 (Nov 6, 2023). Tested on Artifactory 7.71.3 with Terraform CLI v1.6.3

IMPROVEMENTS:

* resource/artifactory_permission_target: Revert back to using Terraform SDKv2 due to unresolved performance issue from Terraform Framework. Issue: [#757](https://github.com/jfrog/terraform-provider-artifactory/issues/757) and [#805](https://github.com/jfrog/terraform-provider-artifactory/issues/805) PR: [#842](https://github.com/jfrog/terraform-provider-artifactory/pull/842)

## 9.7.3 (Nov 2, 2023). Tested on Artifactory 7.71.3 with Terraform CLI v1.6.3

SECURITY:

* provider: Bump google.golang.org/grpc from 1.56.1 to 1.56.3 PR: [#836](https://github.com/jfrog/terraform-provider-artifactory/pull/836)

IMPROVEMENTS:

* resource/artifactory_federated_*_repository: Add configuration synchronization when creating or updating resource. Issue: [#825](https://github.com/jfrog/terraform-provider-artifactory/issues/825)
* provider: Add warning message for Terraform CLI version <1.0.0 deprecation

PR: [#840](https://github.com/jfrog/terraform-provider-artifactory/pull/840)

NOTES:

We will be moving to [Terraform Protocol v6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6) in **Q1 2024**. This means only Terraform CLI version 1.0 and later will be supported.

## 9.7.2 (Oct 19, 2023). Tested on Artifactory 7.71.3 with Terraform CLI v1.6.2

BUG FIXES:

* provider: Fix schema differences between SDKv2 and Framework providers. PR: [#834](https://github.com/jfrog/terraform-provider-artifactory/pull/834) Issue: [#833](https://github.com/jfrog/terraform-provider-artifactory/issues/833)

## 9.7.1 (Oct 19, 2023). Tested on Artifactory 7.68.14 with Terraform CLI v1.6.2

IMPROVEMENTS:

* provider:
  * Remove conflict validation between `access_token` and `api_key` attributes. If set, `access_token` will take precedent over `api_key`.
  * Update documentation to align with actual provider behavior.

PR: [#832](https://github.com/jfrog/terraform-provider-artifactory/pull/832)
Issue: [#663](https://github.com/jfrog/terraform-provider-artifactory/issues/663)

## 9.7.0 (Oct 18, 2023). Tested on Artifactory 7.68.14 with Terraform CLI v1.6.1

IMPROVEMENTS:

* resource/artifactory_keypair is migrated to Plugin Framework. PR: [#829](https://github.com/jfrog/terraform-provider-artifactory/pull/829)
* resource/artifactory_mail_server: remove length validation for optional attribute `subject_prefix`. PR: [#830](https://github.com/jfrog/terraform-provider-artifactory/pull/830) Issue: [#828](https://github.com/jfrog/terraform-provider-artifactory/issues/828)

## 9.6.0 (Oct 13, 2023). Tested on Artifactory 7.68.14 with Terraform CLI v1.6.1

IMPROVEMENTS:

* resource/artifactory_certificate is migrated to Plugin Framework. PR: [#826](https://github.com/jfrog/terraform-provider-artifactory/pull/826)

## 9.5.1 (Oct 12, 2023). Tested on Artifactory 7.68.14 with Terraform CLI v1.6.1

SECURITY:

* provider: Bump golang.org/x/net from 0.11.0 to 0.17.0 PR: [#824](https://github.com/jfrog/terraform-provider-artifactory/pull/824)

## 9.5.0 (Oct 11, 2023). Tested on Artifactory 7.68.14 with Terraform CLI v1.6.1

FEATURES:

* resource/artifactory_local_huggingfaceml_repository, resource/artifactory_remote_huggingfaceml_repository: add new local and remote resources for managing Hugging Face repository. PR: [#823](https://github.com/jfrog/terraform-provider-artifactory/pull/823)

## 9.4.0 (Oct 5, 2023). Tested on Artifactory 7.68.13 with Terraform CLI v1.6.0

FEATURES:

* resource/artifactory_mail_server: add a new resource for managing mail server configuration. PR: [#819](https://github.com/jfrog/terraform-provider-artifactory/pull/819) Issue: [#735](https://github.com/jfrog/terraform-provider-artifactory/issues/735)

## 9.3.1 (Oct 6, 2023). Tested on Artifactory 7.68.13 with Terraform CLI v1.6.0

BUG FIX:
* resource/artifactory_scoped_token: Remove default value for `expires_in` attribute which should fix state drift when upgrading from 7.11.2 or earlier. Issue: [#818](https://github.com/jfrog/terraform-provider-artifactory/issues/818) PR: [#820](https://github.com/jfrog/terraform-provider-artifactory/pull/820)

## 9.3.0 (Oct 3, 2023). Tested on Artifactory 7.68.13 with Terraform CLI v1.5.7

IMPROVEMENTS:

* resource/artifactory_distribution_public_key is migrated to Plugin Framework. PR: [#817](https://github.com/jfrog/terraform-provider-artifactory/pull/817)
* resource/artifactory_remote_\*\_repository: Fix incorrect default value for `store_artifacts_locally` attribute in documentation. PR: [#816](https://github.com/jfrog/terraform-provider-artifactory/pull/816)

## 9.2.1 (Sep 29, 2023). Tested on Artifactory 7.68.11 with Terraform CLI v1.5.7

IMPROVEMENTS:

* Update module path to `/v9` PR: [#814](https://github.com/jfrog/terraform-provider-artifactory/pull/814)

## 9.2.0 (Sep 28, 2023). Tested on Artifactory 7.68.11 with Terraform CLI v1.5.7

IMPROVEMENTS:

* resource/artifactory_backup is migrated to Plugin Framework. PR: [#812](https://github.com/jfrog/terraform-provider-artifactory/pull/812)

## 9.1.0 (Sep 28, 2023). Tested on Artifactory 7.68.11 with Terraform CLI v1.5.7

IMPROVEMENTS:
* data/artifactory_local_conan_repository, data/artifactory_virtual_conan_repository, data/artifactory_federated_conan_repository, resource/artifactory_local_conan_repository, resource/artifactory_virtual_conan_repository, resource/artifactory_federated_conan_repository: add `force_conan_authentication` attribute PR: [#810](https://github.com/jfrog/terraform-provider-artifactory/pull/810) Issue: [#808](https://github.com/jfrog/terraform-provider-artifactory/issues/808)
* resource/artifactory_\*\_repository: update validation for `project_environments` attribute to allow empty list. PR: [#811](https://github.com/jfrog/terraform-provider-artifactory/pull/811)

## 9.0.0 (Sep 15, 2023). Tested on Artifactory 7.68.7 with Terraform CLI v1.5.7

IMPROVEMENTS:
* resource/artifactory_\*\_repository: remove default value of "default" from `project_key` attribute. This is a REST API bug fix that is part of Artifactory v7.68.7 (self-hosted) and v7.67.0 (cloud). Existing Terraform state with "default" value should be automatically migrated to "" on `terraform apply`. No state drift should occurs on `terraform plan`. Issue: [#779](https://github.com/jfrog/terraform-provider-artifactory/issues/779) 
* Fix incorrect description for remote repository attribute `block_mismatching_mime_types`. Issue: [#799](https://github.com/jfrog/terraform-provider-artifactory/pull/799)
* Add multiple users and groups HCL example for `artifactory_permission_target` resource. Issue: [#800](https://github.com/jfrog/terraform-provider-artifactory/pull/800)

PR: [#804](https://github.com/jfrog/terraform-provider-artifactory/pull/804)

## 8.9.1 (Sep 11, 2023). Tested on Artifactory 7.63.14 with Terraform CLI v1.5.7

BUG FIXES:
* resource/artifactory_local_\*\_repository, resource/artifactory_remote_\*\_repository, resource/artifactory_virtual_\*\_repository, resource/artifactory_federated_\*\_repository: fix unable to set `description` and `notes` attributes with empty text.

PR: [#798](https://github.com/jfrog/terraform-provider-artifactory/pull/798)
Issue: [#786](https://github.com/jfrog/terraform-provider-artifactory/issues/786)

## 8.9.0 (Sep 8, 2023). Tested on Artifactory 7.63.14 with Terraform CLI v1.5.7

FEATURES:
* resource/artifactory_global_environment: add a new resource for managing global environment. PR: [#797](https://github.com/jfrog/terraform-provider-artifactory/pull/797) Issue: [#773](https://github.com/jfrog/terraform-provider-artifactory/issues/773)

## 8.8.1 (Sep 7, 2023). Tested on Artifactory 7.63.14 with Terraform CLI v1.5.6

BUG FIXES:
* resource/artifactory_scoped_token: Fix state drift with `description` attribute when upgrading from 7.11.2.

PR: [#796](https://github.com/jfrog/terraform-provider-artifactory/pull/796) Issue: [#792](https://github.com/jfrog/terraform-provider-artifactory/issues/792)

## 8.8.0 (Sep 5, 2023). Tested on Artifactory 7.63.14 with Terraform CLI v1.5.6

IMPROVEMENTS:
* resource/artifactory_user, resource/artifactory_unmanaged_user, resource/artifactory_managed_user: Add validation to `name` attribute to match JFrog platform behavior. PR: [#794](https://github.com/jfrog/terraform-provider-artifactory/pull/794) Issue: [#790](https://github.com/jfrog/terraform-provider-artifactory/issues/790)

## 8.7.0 (August 30, 2023). Tested on Artifactory 7.63.14 with Terraform CLI v1.5.6

IMPROVEMENTS:
* resource/artifactory_remote_helm_repository: Add support of `oci` scheme to `helm_charts_base_url` attribute. PR: [#793](https://github.com/jfrog/terraform-provider-artifactory/pull/793)

## 8.6.0 (August 25, 2023). Tested on Artifactory 7.63.12 with Terraform CLI v1.5.6

IMPROVEMENTS:
* resource/artifactory_remote_docker_repository: Add `disable_url_normalization` attribute to support disabling URL normalization.

PR: [#788](https://github.com/jfrog/terraform-provider-artifactory/pull/788)
Issue: [#784](https://github.com/jfrog/terraform-provider-artifactory/issues/784)

## 8.5.0 (August 23, 2023). Tested on Artifactory 7.63.12 with Terraform CLI v1.5.5

IMPROVEMENTS:
* resource/artifactory_scoped_token: Add `project_key` attribute to support creating token for a project.

PR: [#787](https://github.com/jfrog/terraform-provider-artifactory/pull/787)

## 8.4.3. (August 23, 2023). Tested on Artifactory 7.63.12 with Terraform CLI v1.5.5

IMPROVEMENTS: 
* resource/artifactory_permission_target: added 0 length check for `actions` block. 
* Upgraded Terraform Plugin Framework to 1.3.5. 

PR: [#785](https://github.com/jfrog/terraform-provider-artifactory/pull/785)
Issues: [#782](https://github.com/jfrog/terraform-provider-artifactory/issues/782)

## 8.4.2 (August 22, 2023). Tested on Artifactory 7.63.12 with Terraform CLI v1.5.5

IMPROVEMENTS:
* resource/artifactory_access_token: Add missing deprecation message to documentation. The message has been part of the provider since [6.8.0](https://github.com/jfrog/terraform-provider-artifactory/releases/tag/v6.8.0) but was missing from the documentation on Terraform registry.

PR: [#783](https://github.com/jfrog/terraform-provider-artifactory/pull/783)

## 8.4.1 (July 17, 2023). Tested on Artifactory 7.63.12 with Terraform CLI v1.5.5

BUG FIXES:
* resource/artifactory_group, resource/artifactory_permission_target, resource/artifactory_scoped_token, resource/artifactory_ldap_setting_v2, resource/artifactory_ldap_group_setting_v2, resource/artifactory_*_user: fixed handling of HTTP response 404. When the resource was removed outside of Terraform configuration, it will be recreated without error out on 404. 

PR: [#781](https://github.com/jfrog/terraform-provider-artifactory/pull/781)
Issues: [#776](https://github.com/jfrog/terraform-provider-artifactory/issues/776), [#775](https://github.com/jfrog/terraform-provider-artifactory/issues/775), [#772](https://github.com/jfrog/terraform-provider-artifactory/issues/772)

## 8.4.0 (July 21, 2023). Tested on Artifactory 7.63.5 with Terraform CLI v1.5.3

IMPROVEMENTS:
* resource/artifactory_ldap_setting_v2 and resource/artifactory_ldap_group_setting_v2 were added to replace `artifactory_ldap_setting` and `artifactory_ldap_group_setting`. The new resources uses a new API access endpoint, introduced in Artifactory 7.57.1 and will work on both self-hosted and SaaS versions starting from 7.57.1 and above. 
 Older resources will still work on self-hosted Artifactory versions below 7.57.1.
 [LDAP](https://jfrog.com/help/r/jfrog-rest-apis/ldap-setting?page=40) and [LDAP group](https://jfrog.com/help/r/jfrog-rest-apis/ldap-group-setting) API documentation. 
* resource/artifact_webhook, resource/artifact_property_webhook, resource/artifactory_release_bundle_webhook, resource/build_webhook, resource/distribution_webhook, resource/docker_webhook, resource/release_bundle_webhook changed the way we are unpacking `secret` attribute, due to the change in the Artifactory API. Secret is hashed now in the GET call response, and we need to use the value, saved in the TF state. 

PR:[#769](https://github.com/jfrog/terraform-provider-artifactory/pull/769), [#770](https://github.com/jfrog/terraform-provider-artifactory/pull/770)
Issues:[#768](https://github.com/jfrog/terraform-provider-artifactory/issues/768), [#749](https://github.com/jfrog/terraform-provider-artifactory/issues/749)

## 8.3.1 (July 5, 2023). Tested on Artifactory 7.59.12 with Terraform CLI v1.5.2

BUG FIXES:
* resource/artifactory_scoped_token: default value `false` is removed from `include_reference_token` attribute to avoid state drift, when the provider updating from version below 7.7.0 to the latest. 

PR: [#763](https://github.com/jfrog/terraform-provider-artifactory/pull/763)
Issues:[#758](https://github.com/jfrog/terraform-provider-artifactory/issues/758), [#761](https://github.com/jfrog/terraform-provider-artifactory/issues/761)

## 8.3.0 (July 5, 2023). Tested on Artifactory 7.59.11 with Terraform CLI v1.5.2

IMPROVEMENTS:

* New resources added: resource/artifact_custom_webhook, resource/artifact_property_custom_webhook, resource/artifactory_release_bundle_custom_webhook, resource/build_custom_webhook, resource/distribution_custom_webhook, resource/docker_custom_webhook, resource/release_bundle_custom_webhook. These new resource allow to create custom webhooks. [API documentation](https://jfrog.com/help/r/jfrog-rest-apis/create-a-new-webhook-subscription?page=38), [Custom Webhooks documentation](https://jfrog.com/help/r/jfrog-platform-administration-documentation/custom-webhooks?page=3). 

PR:[#762](https://github.com/jfrog/terraform-provider-artifactory/pull/762)
Issue:[#738](https://github.com/jfrog/terraform-provider-artifactory/issues/738)


## 8.2.3 (June 28, 2023). Tested on Artifactory 7.59.11 with Terraform CLI v1.5.2

BUG FIXES:
* provider: fixed "Error: Plugin did not respond" issue. Traced to a [Terraform framework library bug](https://github.com/hashicorp/terraform-plugin-framework/pull/772). Updated `terraform-plugin-framework` from v1.3.0 to [v1.3.1](https://github.com/hashicorp/terraform-plugin-framework/releases/tag/v1.3.1)

  PR: [#759](https://github.com/jfrog/terraform-provider-artifactory/pull/759)
  Issue: [#757](https://github.com/jfrog/terraform-provider-artifactory/issues/757)

## 8.2.2 (June 22, 2023). Tested on Artifactory 7.59.11 with Terraform CLI v1.5.1

BUG FIXES:
* resource/artifactory_remote_vcs_repository: IsURLWithHTTPorHTTPS check was removed from the `vcs_git_download_url` attribute to allow use formatted string with placeholders, like `{0}/{1}/+archive/{2}.{3}` with `CUSTOM` Git provider.

  PR: [#756](https://github.com/jfrog/terraform-provider-artifactory/pull/756)
  Issue: [#747](https://github.com/jfrog/terraform-provider-artifactory/issues/747)

## 8.2.1 (June 21, 2023). Tested on Artifactory 7.59.11 with Terraform CLI v1.5.1

IMPROVEMENTS:

* resource/artifactory_remote_*_repository: changed behavior of attribute `remote_repo_layout_ref` to match the UI behavior. User still can create a repo without that attribute, but once it's set, it can't be removed or updated with an empty string. 

  PR: [#755](https://github.com/jfrog/terraform-provider-artifactory/pull/755)
  Issue: [#746](https://github.com/jfrog/terraform-provider-artifactory/issues/746)

## 8.2.0 (June 19, 2023). Tested on Artifactory 7.59.11 with Terraform CLI v1.5.0

IMPROVEMENTS:

* resource/artifactory_remote_*_repository: new attribute `disable_proxy` was added. When set to `true`, proxy settings are ignored for the remote repository.  
  
  PR:    [#751](https://github.com/jfrog/terraform-provider-artifactory/pull/751)
  Issue: [#739](https://github.com/jfrog/terraform-provider-artifactory/issues/739)

## 8.1.0 (June 15, 2023). Tested on Artifactory 7.59.9

NOTES:

Due to the complexity of maintaining backward compatibility when migrating from Terraform SDKv2 to plugin-framework, we inadvertently changed these resources schema without the necessary acceptance test coverage. With help from HashiCorp Terraform team, we think we have fixed these issues and maintained backward compatibility.

However, if you have created new resources using these resource types after they were initially migrated, you may now have state drift when you upgrade to this version. This is caused by the fix in this version, which revert the resource schema to mimic behavior from pre-migration provider.

BUG FIX:
* resource/artifactory_group, resource/artifactory_scoped_token, resource/artifactory_managed_user, resource/artifactory_user, resource_permission_target: Fixed unintended state drift when upgrading from pre-migrated provider.

PR: [#750](https://github.com/jfrog/terraform-provider-artifactory/pull/750)
Issues: [#744](https://github.com/jfrog/terraform-provider-artifactory/issues/744)

## 8.0.0 (June 1, 2023). Tested on Artifactory 7.59.9

BREAKING CHANGES:

* resource/artifactory_permission_targets has been removed. It has been marked as deprecated since April 2022.

IMPROVEMENTS: 

* resource/artifactory_permission_target is migrated to Plugin Framework, improved attribute validation.

PR: [#742](https://github.com/jfrog/terraform-provider-artifactory/pull/742)

## 7.11.2 (May 30, 2023). Tested on Artifactory 7.59.9

IMPROVEMENTS:

* resource/scoped_token is migrated to Plugin Framework, improved attribute validation.

PR: [#741](https://github.com/jfrog/terraform-provider-artifactory/pull/741)

NOTES:

Resources `ldap_group_setting` and `ldap_setting` won't work with Artifactory version => 7.57.1. 
The new API was implemented to manage LDAP configuration and the new resource will be added later. 

## 7.11.1 (May 23, 2023). Tested on Artifactory 7.55.14

IMPROVEMENTS: 

* resource/artifactory_local_repository_single_replication: the resource can deal with different license types (Enterprise and ProX) to create replications. 
The reason the change introduced, is the API response body is different for different license types.

PR:    [#737](https://github.com/jfrog/terraform-provider-artifactory/pull/737)
Issue: [#718](https://github.com/jfrog/terraform-provider-artifactory/issues/718)

## 7.11.0 (May 16, 2023). Tested on Artifactory 7.55.13

IMPROVEMENTS:

* resource/artifactory_group is migrated to Plugin Framework, improved attribute validation.

  PR:     [#734](https://github.com/jfrog/terraform-provider-artifactory/pull/734)

BUG FIXES: 

* fixed the issue when nil pointer happens in some cases if JFROG_URL is not set. 

  Issues: [#731](https://github.com/jfrog/terraform-provider-artifactory/issues/731)


## 7.10.1 (May 10, 2023).

BUG FIXES: 
* Fixed bug where `check_license` attribute was always `true` in the SDK v2 provider configuration.

  PR:     [#733](https://github.com/jfrog/terraform-provider-artifactory/pull/733)
  Issues: [#732](https://github.com/jfrog/terraform-provider-artifactory/issues/732)


## 7.10.0 (May 8, 2023).

BUG FIXES:
* Fixed a diff between SDK v2 and Plugin Framework providers schemas, which created a problem during the update process from older versions to 7.8.0
* Removed default functions for the provider schema attributes. Defaults are set in the configuration step now to avoid schemas conflicts.

  PR:     [#730](https://github.com/jfrog/terraform-provider-artifactory/pull/730)
  Issues: 
  * [#728](https://github.com/jfrog/terraform-provider-artifactory/issues/728)
  * [#729](https://github.com/jfrog/terraform-provider-artifactory/issues/729)

## 7.9.0 (May 8, 2023). Tested on Artifactory 7.55.10

FEATURES:

* resource/artifactory_distribution_public_key: Adds new resource to manage distribution public keys which are used to verify signed release bundles
  PR:     [#725](https://github.com/jfrog/terraform-provider-artifactory/pull/725)
  Issues: [#721](https://github.com/jfrog/terraform-provider-artifactory/issues/721)

## 7.8.0 (May 5, 2023). Tested on Artifactory 7.55.10

IMPROVEMENTS:

* Start of the migration from SDK v2 to Terraform Plugin Framework.
* added provider muxing.
* resource/artifactory_user, resource/artifactory_anonymous_user and resource/artifactory_managed_user migrated to the framework.
* added templates and examples for auto-generated documentation for users.

PR [#726](https://github.com/jfrog/terraform-provider-artifactory/pull/726)

## 7.7.0 (April 26, 2023). Tested on Artifactory 7.55.10

IMPROVEMENTS:

* resource/artifactory_scoped_token: adds reference_token and include_reference_token attributes for the resource.
  PR [#723](https://github.com/jfrog/terraform-provider-artifactory/pull/723)

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

* resource/artifactory_scoped_token: Add `Sensitive: true` to `access_token` and `refresh_token` attributes to ensure the values are handled correctly.

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
