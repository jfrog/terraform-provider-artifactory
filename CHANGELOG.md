## 2.16.2 (Feb, 10, 2022)

BUG FIXES:

* resource/artifactory_single_replication_config: Fix for error when repository got externally removed, but replication resource configured. [GH-312]

## 2.16.1 (Feb, 7, 2022)

BUG FIXES:

* resource/artifactory_remote_repository: Fix failing test for `proxy` attribute [GH-311]

## 2.16.0 (Feb, 4, 2022)

IMPROVEMENTS:

* resource/artifactory_group: Added support for manager roles in artifactory_group resource [GH-308]

## 2.15.2 (Feb, 4, 2022)

BUG FIXES:

* resource/artifactory_remote_repository: Fix unable to reset `proxy` attribute [GH-307]

## 2.15.1 (Feb, 4, 2022)

BUG FIXES:

* resource/artifactory_xray_watch: Fix incorrect usage of variable reference with Resty `.SetBody()` in `create` and `update` funcs [GH-306]

## 2.15.0 (Feb, 3, 2022)

FEATURES:

* **New Resource:** `artifactory_virtual_rpm_repository` with support for `primary_keypair_ref` and `secondary_keypair_ref` and [GH-303]

## 2.14.0 (Feb, 3, 2022)

FEATURES:

* Added following smart remote repo attributes for npm, cargo, docker and helm remote repository resources [GH-305].
  * "statistics_enabled"
  * "properties_enabled"
  * "source_origin_absence_detection"

## 2.13.1 (Feb, 2, 2022)

IMPROVEMENTS:

* Add missing documentations for Federated repo resources [GH-304]
* Add additional repo types for Federated repo resources [GH-304]

## 2.13.0 (Feb, 1, 2022)

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
