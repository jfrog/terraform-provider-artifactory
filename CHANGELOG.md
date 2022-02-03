## 2.14.0 (Feb, 2, 2022)

IMPROVEMENTS:

* Add missing documentations for Federated repo resources [GH-304]
* Add additional repo types for Federated repo resources [GH-304]
* Added following smart remote repo attributes for remote repository resources [GH-305].
  * "statistics_enabled"
  * "properties_enabled"
  * "source_origin_absence_detection"
  
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
