## 2.23.2 (Mar 17, 2022)

IMPROVEMENTS:

* Datasource `datasource_artifactory_file`, added a parameter `dereference`, 
  which allows users to download the latest artifacts from repositories [GH-361]

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

*Conditional file download depending on `force_overwrite` value of data source `artifactory_file`. [GH-352]

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
