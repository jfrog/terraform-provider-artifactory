# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "7.0.2"
    }
  }
}

provider "artifactory" {
  //  supply ARTIFACTORY_ACCESS_TOKEN / JFROG_ACCESS_TOKEN / ARTIFACTORY_API_KEY and ARTIFACTORY_URL / JFROG_URL as env vars
}

resource "artifactory_user" "new_user" {
  name   = "new_user"
  email  = "new_user@somewhere.com"
  groups = ["readers"]
}

resource "artifactory_scoped_token" "user" {
  username = artifactory_user.new_user.name
}

resource "artifactory_scoped_token" "admin" {
  scopes = ["applied-permissions/admin"]
}

output "user-scoped-token" {
  value = artifactory_scoped_token.user
}

output "admin-scoped-token" {
  value = artifactory_scoped_token.admin
}

resource "artifactory_local_bower_repository" "bower-local" {
  key         = "bower-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_cargo_repository" "cargo-local" {
  key         = "cargo-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_chef_repository" "chef-local" {
  key         = "chef-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_cocoapods_repository" "cocoapods-local" {
  key         = "cocoapods-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_composer_repository" "composer-local" {
  key         = "composer-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_conan_repository" "conan-local" {
  key         = "conan-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_conda_repository" "conda-local" {
  key         = "conda-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_cran_repository" "cran-local" {
  key         = "cran-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_docker_v2_repository" "docker-v2-local" {
  key             = "docker-v2-local"
  tag_retention   = 3
  max_unique_tags = 5
}

resource "artifactory_local_docker_v1_repository" "docker-v1-local" {
  key         = "docker-v1-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_gems_repository" "gems-local" {
  key         = "gems-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_generic_repository" "generic-local" {
  key         = "generic-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_gitlfs_repository" "gitlfs-local" {
  key         = "gitlfs-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_go_repository" "go-local" {
  key         = "go-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_gradle_repository" "my-gradle-local" {
  key                             = "my-gradle-local"
  checksum_policy_type            = "client-checksums"
  snapshot_version_behavior       = "unique"
  max_unique_snapshots            = 10
  handle_releases                 = true
  handle_snapshots                = true
  suppress_pom_consistency_checks = true
}

resource "artifactory_local_helm_repository" "helm-local" {
  key         = "helm-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_ivy_repository" "ivy-local" {
  key         = "ivy-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_maven_repository" "maven-local" {
  key                             = "maven-local"
  checksum_policy_type            = "client-checksums"
  snapshot_version_behavior       = "unique"
  max_unique_snapshots            = 10
  handle_releases                 = true
  handle_snapshots                = true
  suppress_pom_consistency_checks = false
}

resource "artifactory_local_npm_repository" "my-npm-local" {
  key         = "my-npm-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_nuget_repository" "my-nuget-local" {
  key                        = "my-nuget-local"
  max_unique_snapshots       = 10
  force_nuget_authentication = true
}

resource "artifactory_local_opkg_repository" "opkg-local" {
  key         = "opkg-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_pub_repository" "pub-local" {
  key         = "pub-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_puppet_repository" "puppet-local" {
  key         = "puppet-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_pypi_repository" "pypi-local" {
  key         = "pypi-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_rpm_repository" "rpm-local" {
  key                        = "rpm-local"
  yum_root_depth             = 5
  calculate_yum_metadata     = true
  enable_file_lists_indexing = true
  yum_group_file_names       = "file-1.xml,file-2.xml"
}

resource "artifactory_local_sbt_repository" "sbt-local" {
  key         = "sbt-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_terraform_module_repository" "terraform-module-local" {
  key         = "terraform-module-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_terraform_provider_repository" "terraform-provider-local" {
  key         = "terraform-provider-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_terraformbackend_repository" "terraformbackend-local" {
  key         = "terraformbackend-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "artifactory_local_vagrant_repository" "vagrant-local" {
  key         = "vagrant-local"
  description = "Repo created by Terraform Provider Artifactory"
}

resource "random_id" "randid" {
  byte_length = 16
}

resource "tls_private_key" "example" {
  algorithm = "RSA"
  rsa_bits  = 2048

}
resource "artifactory_keypair" "some-keypairRSA" {
  pair_name   = "some-keypairfoo"
  pair_type   = "RSA"
  private_key = file("samples/rsa.priv")
  public_key  = file("samples/rsa.pub")
  alias       = "foo-aliasfoo"
  passphrase  = "some-passphrase"
}

resource "artifactory_keypair" "some-keypairGPG1" {
  pair_name   = "some-keypair${random_id.randid.id}"
  pair_type   = "GPG"
  alias       = "foo-alias1"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")
  passphrase  = "some-passphrase"
}

resource "artifactory_keypair" "some-keypairGPG2" {
  pair_name   = "some-keypair4${random_id.randid.id}"
  pair_type   = "GPG"
  alias       = "foo-alias2"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")
  passphrase  = "some-passphrase"
}

resource "artifactory_local_debian_repository" "my-debian-repo" {
  key                       = "my-debian-repo"
  primary_keypair_ref       = artifactory_keypair.some-keypairGPG1.pair_name
  secondary_keypair_ref     = artifactory_keypair.some-keypairGPG2.pair_name
  index_compression_formats = ["bz2", "lzma", "xz"]
  depends_on                = [artifactory_keypair.some-keypairGPG1, artifactory_keypair.some-keypairGPG2]
}

resource "artifactory_local_alpine_repository" "terraform-local-test-repo-basic1896042683811651651" {
  key                 = "terraform-local-test-repo-basic1896042683811651651"
  primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  depends_on          = [artifactory_keypair.some-keypairRSA]
}

resource "artifactory_remote_alpine_repository" "my-remote-alpine" {
  key = "my-remote-alpine"
  url = "http://dl-cdn.alpinelinux.org/alpine"
}

resource "artifactory_remote_bower_repository" "my-remote-bower" {
  key              = "my-remote-bower"
  url              = "https://github.com/"
  vcs_git_provider = "GITHUB"
}

resource "artifactory_remote_cargo_repository" "my-remote-cargo" {
  key              = "my-remote-cargo"
  anonymous_access = true
  url              = "https://github.com/"
  git_registry_url = "https://github.com/rust-lang/foo.index"
}

resource "artifactory_remote_chef_repository" "my-remote-chef" {
  key = "my-remote-chef"
  url = "https://supermarket.chef.io"
}

resource "artifactory_remote_cocoapods_repository" "my-remote-cocoapods" {
  key              = "my-remote-cocoapods"
  url              = "https://github.com/"
  vcs_git_provider = "GITHUB"
}

resource "artifactory_remote_composer_repository" "my-remote-composer" {
  key              = "my-remote-composer"
  url              = "https://github.com/"
  vcs_git_provider = "GITHUB"
}

resource "artifactory_remote_conan_repository" "my-remote-conan" {
  key                        = "my-remote-conan"
  url                        = "https://conan.bintray.com"
  force_conan_authentication = true
}

resource "artifactory_remote_conda_repository" "my-remote-conda" {
  key = "my-remote-conda"
  url = "https://repo.anaconda.com/pkgs/main"
}

resource "artifactory_remote_cran_repository" "my-remote-cran" {
  key = "my-remote-cran"
  url = "https://cran.r-project.org/"
}

resource "artifactory_remote_debian_repository" "my-remote-debian" {
  key = "my-remote-Debian"
  url = "http://archive.ubuntu.com/ubuntu/"
}

resource "artifactory_remote_docker_repository" "my-remote-docker" {
  key                            = "my-remote-docker"
  external_dependencies_enabled  = true
  external_dependencies_patterns = ["**/registry-1.docker.io/**"]
  enable_token_authentication    = true
  url                            = "https://registry-1.docker.io/"
  block_pushing_schema1          = true
}

resource "artifactory_remote_gems_repository" "my-remote-gems" {
  key = "my-remote-gems"
  url = "https://rubygems.org/"
}

resource "artifactory_remote_generic_repository" "my-remote-generic" {
  key = "my-remote-generic"
  url = "http://testartifactory.io/artifactory/example-generic/"
}

resource "artifactory_remote_gitlfs_repository" "my-remote-gitlfs" {
  key = "my-remote-gitlfs"
  url = "http://testartifactory.io/artifactory/example-gitlfs/"
}

resource "artifactory_remote_go_repository" "my-remote-go" {
  key              = "my-remote-go"
  url              = "https://proxy.golang.org/"
  vcs_git_provider = "ARTIFACTORY"
}

resource "artifactory_remote_gradle_repository" "gradle-remote" {
  key                             = "gradle-remote-foo"
  url                             = "https://repo1.maven.org/maven2/"
  fetch_jars_eagerly              = true
  fetch_sources_eagerly           = false
  suppress_pom_consistency_checks = true
  reject_invalid_jars             = true
}

resource "artifactory_remote_helm_repository" "helm-remote" {
  key                           = "helm-remote-foo25"
  url                           = "https://repo.chartcenter.io/"
  helm_charts_base_url          = "https://foo.com"
  external_dependencies_enabled = true
  external_dependencies_patterns = [
    "**github.com**"
  ]
}

resource "artifactory_remote_ivy_repository" "ivy-remote" {
  key                             = "ivy-remote-foo"
  url                             = "https://repo1.maven.org/maven2/"
  fetch_jars_eagerly              = true
  fetch_sources_eagerly           = false
  suppress_pom_consistency_checks = true
  reject_invalid_jars             = true
}

resource "artifactory_remote_maven_repository" "maven-remote" {
  key                                = "maven-remote-foo"
  url                                = "https://repo1.maven.org/maven2/"
  fetch_jars_eagerly                 = true
  fetch_sources_eagerly              = false
  suppress_pom_consistency_checks    = false
  reject_invalid_jars                = true
  metadata_retrieval_timeout_seconds = 120
}

resource "artifactory_remote_npm_repository" "thing" {
  key                                  = "remote-thing-npm"
  url                                  = "https://registry.npmjs.org/"
  list_remote_folder_items             = true
  mismatching_mime_types_override_list = "application/json,application/xml"
  xray_index                           = true
}

resource "artifactory_remote_nuget_repository" "my-remote-nuget" {
  key                        = "my-remote-nuget"
  url                        = "https://www.nuget.org/"
  download_context_path      = "api/v2/package"
  force_nuget_authentication = true
  v3_feed_url                = "https://api.nuget.org/v3/index.json"
}

resource "artifactory_remote_opkg_repository" "my-remote-opkg" {
  key = "my-remote-opkg"
  url = "http://testartifactory.io/artifactory/example-opkg/"
}

resource "artifactory_remote_p2_repository" "my-remote-p2" {
  key = "my-remote-p2"
  url = "http://testartifactory.io/artifactory/example-p2/"
}

resource "artifactory_remote_pub_repository" "my-remote-pub" {
  key = "my-remote-pub"
  url = "https://pub.dartlang.org"
}

resource "artifactory_remote_puppet_repository" "my-remote-puppet" {
  key = "my-remote-puppet"
  url = "https://forgeapi.puppetlabs.com/"
}

resource "artifactory_remote_pypi_repository" "pypi_remote" {
  key               = "pypi-remote"
  url               = "https://files.pythonhosted.org"
  pypi_registry_url = "https://custom.PYPI.registry.url"
}

resource "artifactory_remote_terraform_repository" "terraform-remote" {
  key                     = "terraform-remote"
  url                     = "https://github.com/"
  terraform_registry_url  = "https://registry.terraform.io"
  terraform_providers_url = "https://releases.hashicorp.com"
}

resource "artifactory_remote_rpm_repository" "my-remote-rpm" {
  key = "my-remote-rpm"
  url = "http://mirror.centos.org/centos/"
}

resource "artifactory_remote_sbt_repository" "sbt-remote" {
  key                             = "sbt-remote-foo"
  url                             = "https://repo1.maven.org/maven2/"
  fetch_jars_eagerly              = true
  fetch_sources_eagerly           = false
  suppress_pom_consistency_checks = true
  reject_invalid_jars             = true
}

resource "artifactory_remote_vcs_repository" "my-remote-vcs" {
  key                  = "my-remote-vcs"
  url                  = "https://github.com/"
  vcs_git_provider     = "GITHUB"
  max_unique_snapshots = 5
}

resource "artifactory_virtual_alpine_repository" "foo-alpine" {
  key              = "foo-alpine"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_bower_repository" "foo-bower" {
  key                           = "foo-bower"
  repositories                  = []
  description                   = "A test virtual repo"
  notes                         = "Internal description"
  includes_pattern              = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern              = "com/google/**"
  external_dependencies_enabled = false
}

resource "artifactory_virtual_chef_repository" "foo-chef" {
  key              = "foo-chef"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_composer_repository" "foo-composer" {
  key              = "foo-composer"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_conan_repository" "foo-conan" {
  key              = "foo-conan"
  repo_layout_ref  = "conan-default"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_conda_repository" "foo-conda" {
  key              = "foo-conda"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_cran_repository" "foo-cran" {
  key              = "foo-cran"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_debian_repository" "foo-debian" {
  key                                = "foo-debian"
  repositories                       = []
  description                        = "A test virtual repo"
  notes                              = "Internal description"
  includes_pattern                   = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                   = "com/google/**"
  optional_index_compression_formats = ["bz2", "xz"]
  debian_default_architectures       = "amd64,i386"
}

resource "artifactory_virtual_docker_repository" "foo-docker" {
  key              = "foo-docker"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_gems_repository" "foo-gems" {
  key              = "foo-gems"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_generic_repository" "foo-generic" {
  key              = "foo-generic"
  repo_layout_ref  = "simple-default"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_gitlfs_repository" "foo-gitlfs" {
  key              = "foo-gitlfs"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_go_repository" "baz-go" {
  key                           = "baz-go"
  repo_layout_ref               = "go-default"
  repositories                  = []
  description                   = "A test virtual repo"
  notes                         = "Internal description"
  includes_pattern              = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern              = "com/google/**"
  external_dependencies_enabled = true
  external_dependencies_patterns = [
    "**/github.com/**",
    "**/go.googlesource.com/**"
  ]
}

resource "artifactory_virtual_gradle_repository" "foo-gradle" {
  key                                      = "foo-gradle"
  repositories                             = []
  description                              = "A test virtual repo"
  notes                                    = "Internal description"
  includes_pattern                         = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                         = "com/google/**"
  pom_repository_references_cleanup_policy = "discard_active_reference"
}

resource "artifactory_virtual_helm_repository" "foo-helm-virtual" {
  key            = "foo-helm-virtual"
  use_namespaces = true
}

resource "artifactory_virtual_ivy_repository" "foo-ivy" {
  key                                      = "foo-ivy"
  repositories                             = []
  description                              = "A test virtual repo"
  notes                                    = "Internal description"
  includes_pattern                         = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                         = "com/google/**"
  pom_repository_references_cleanup_policy = "discard_active_reference"
}

resource "artifactory_virtual_maven_repository" "foo" {
  key             = "maven-virt-repo"
  repo_layout_ref = "maven-2-default"
  repositories = [
    artifactory_local_maven_repository.maven-local.key,
    artifactory_remote_maven_repository.maven-remote.key
  ]
  description                              = "A test virtual repo"
  notes                                    = "Internal description"
  includes_pattern                         = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                         = "com/google/**"
  force_maven_authentication               = true
  pom_repository_references_cleanup_policy = "discard_active_reference"
}

resource "artifactory_virtual_npm_repository" "foo-npm" {
  key              = "foo-npm"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_nuget_repository" "foo-nuget" {
  key                        = "foo-nuget"
  repositories               = []
  description                = "A test virtual repo"
  notes                      = "Internal description"
  includes_pattern           = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern           = "com/google/**"
  force_nuget_authentication = true
}

resource "artifactory_virtual_p2_repository" "foo-p2" {
  key              = "foo-p2"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_pub_repository" "foo-pub" {
  key              = "foo-pub"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_puppet_repository" "foo-puppet" {
  key              = "foo-puppet"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_pypi_repository" "foo-pypi" {
  key              = "foo-pypi"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}

resource "artifactory_virtual_rpm_repository" "foo-rpm-virtual" {
  key = "foo-rpm-virtual"

  primary_keypair_ref   = artifactory_keypair.some-keypairGPG1.pair_name
  secondary_keypair_ref = artifactory_keypair.some-keypairGPG2.pair_name

  depends_on = [
    artifactory_keypair.some-keypairGPG1,
    artifactory_keypair.some-keypairGPG2,
  ]
}

resource "artifactory_virtual_sbt_repository" "foo-sbt" {
  key                                      = "foo-sbt"
  repositories                             = []
  description                              = "A test virtual repo"
  notes                                    = "Internal description"
  includes_pattern                         = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                         = "com/google/**"
  pom_repository_references_cleanup_policy = "discard_active_reference"
}

resource "artifactory_virtual_terraform_repository" "terraform-virtual" {
  key              = "terraform-remote"
  repositories     = []
  description      = "A test virtual repo"
  notes            = "Internal description"
  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern = "com/google/**"
}


resource "artifactory_federated_generic_repository" "generic-federated-1" {
  key = "generic-federated-1"

  member {
    url     = "http://artifactory-2:8081/artifactory/federated-generic-5"
    enabled = true
  }

  member {
    url     = "http://artifactory-2:8081/artifactory/federated-generic-6"
    enabled = true
  }
}