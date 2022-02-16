# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.19.1"
    }
  }
}

provider "artifactory" {
  //  supply ARTIFACTORY_USERNAME, _PASSWORD and _URL as env vars
}

resource "artifactory_local_nuget_repository" "my-nuget-local" {
  key                        = "my-nuget-local"
  max_unique_snapshots       = 10
  force_nuget_authentication = true
}

resource "artifactory_local_generic_repository" "my-generic-local" {
  key                 = "my-generic-local"
}

resource "artifactory_local_npm_repository" "my-npm-local" {
  key                 = "my-npm-local"
}

resource "artifactory_local_maven_repository" "my-maven-local" {
  key                             = "my-maven-local"
  checksum_policy_type            = "client-checksums"
  snapshot_version_behavior       = "unique"
  max_unique_snapshots            = 10
  handle_releases                 = true
  handle_snapshots                = true
  suppress_pom_consistency_checks = false
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

resource "artifactory_local_docker_v2_repository" "foo" {
  key             = "foo"
  tag_retention   = 3
  max_unique_tags = 5
}

resource "artifactory_local_docker_v1_repository" "foo" {
  key = "foo"
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
  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_keypair" "some-keypairGPG1" {
  pair_name   = "some-keypair${random_id.randid.id}"
  pair_type   = "GPG"
  alias       = "foo-alias1"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")
  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_keypair" "some-keypairGPG2" {
  pair_name   = "some-keypair4${random_id.randid.id}"
  pair_type   = "GPG"
  alias       = "foo-alias2"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")
  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_local_debian_repository" "my-debian-repo" {
  key                       = "my-debian-repo"
  primary_keypair_ref       = artifactory_keypair.some-keypairGPG1.pair_name
  secondary_keypair_ref     = artifactory_keypair.some-keypairGPG2.pair_name
  index_compression_formats = ["bz2", "lzma", "xz"]
  trivial_layout            = true
  depends_on                = [artifactory_keypair.some-keypairGPG1, artifactory_keypair.some-keypairGPG2]
}

resource "artifactory_local_alpine_repository" "terraform-local-test-repo-basic1896042683811651651" {
  key                 = "terraform-local-test-repo-basic1896042683811651651"
  primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  depends_on          = [artifactory_keypair.some-keypairRSA]
}

variable "supported_repo_types" {
  type = list(string)
  default = [
    "alpine",
    "bower",
    // xray refuses to cargo. They also require a mandatory field we can't currently support
    //    "cargo",
    "chef",
    "cocoapods",
    "composer",
    "conan",
    "conda",
    "cran",
    "debian",
    "docker",
    "gems",
    "generic",
    "gitlfs",
    "go",
    "gradle",
    "helm",
    "ivy",
    "maven",
    "npm",
    "nuget",
    "opkg",
    "p2",
    "puppet",
    "pypi",
    // type 'yum' is not to be supported, as this is really of type 'rpm'. When 'yum' is used on create, RT will
    // respond with 'rpm' and thus confuse TF into think there has been a state change.
    "rpm",
    "sbt",
    "vagrant",
    "vcs",
  ]
}

resource "artifactory_local_repository" "local" {
  count        = length(var.supported_repo_types)
  key          = "${var.supported_repo_types[count.index]}-local"
  package_type = var.supported_repo_types[count.index]
  xray_index   = false
  description  = "hello ${var.supported_repo_types[count.index]}-local"
}

resource "artifactory_local_repository" "local-rand" {
  count        = 100
  key          = "foo-${count.index}-local"
  package_type = var.supported_repo_types[random_id.randid.dec % length(var.supported_repo_types)]
  xray_index   = true
  description  = "hello ${count.index}-local"
}

resource "artifactory_remote_repository" "npm-remote" {
  key          = "npm-remote"
  package_type = "npm"
  url          = "https://registry.npmjs.org"
  xray_index   = true
}

resource "artifactory_remote_pypi_repository" "pypi_remote" {
  key               = "pypi-remote"
  url               = "https://files.pythonhosted.org"
  pypi_registry_url = "https://custom.PYPI.registry.url"
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

resource "artifactory_remote_npm_repository" "thing" {
  key                         = "remote-thing-npm"
  url                         = "https://registry.npmjs.org/"
  repo_layout_ref             = "npm-default"
  missed_cache_period_seconds = 0
  list_remote_folder_items    = true
  mismatching_mime_types_override_list = "application/json,application/xml"
}

resource "artifactory_virtual_maven_repository" "foo" {
  key                                      = "maven-virt-repo"
  repo_layout_ref                          = "maven-2-default"
  repositories                             = []
  description                              = "A test virtual repo"
  notes                                    = "Internal description"
  includes_pattern                         = "com/jfrog/**,cloud/jfrog/**"
  excludes_pattern                         = "com/google/**"
  force_maven_authentication               = true
  pom_repository_references_cleanup_policy = "discard_active_reference"
}

resource "artifactory_virtual_rpm_repository" "rpm-virt-repo" {
  key                   = "rpm-virt-repo"

  primary_keypair_ref   = artifactory_keypair.some-keypairGPG1.pair_name
  secondary_keypair_ref = artifactory_keypair.some-keypairGPG2.pair_name

  depends_on = [
    artifactory_keypair.some-keypairGPG1,
    artifactory_keypair.some-keypairGPG2,
  ]
}

resource "artifactory_federated_generic_repository" "generic-federated-1" {
  key = "generic-federated-1"

  member {
    url    = "http://tempurl.org/artifactory/generic-federated-1"
    enabled = true
  }

  member {
    url    = "http://tempurl2.org/artifactory/generic-federated-2"
    enabled = true
  }
}

resource "artifactory_artifact_webhook" "artifact-webhook" {
  key = "artifact-webhook"
  event_types = ["deployed", "deleted", "moved", "copied"]
  criteria {
    any_local = true
    any_remote = false
    repo_keys = [artifactory_local_repository.local.key]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }

  depends_on = [artifactory_local_repository.local]
}

resource "artifactory_artifact_property_webhook" "artifact-property-webhook" {
  key = "artifact-prop-webhook"
  event_types = ["added", "deleted"]
  criteria {
    any_local = true
    any_remote = false
    repo_keys = [artifactory_local_repository.local.key]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }

  depends_on = [artifactory_local_repository.local]
}

resource "artifactory_docker_webhook" "docker-webhook" {
  key = "docker-webhook"
  event_types = ["pushed", "deleted", "promoted"]
  criteria {
    any_local = true
    any_remote = false
    repo_keys = [artifactory_local_docker_v2_repository.foo.key]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }

  depends_on = [artifactory_local_docker_v2_repository.foo]
}

resource "artifactory_build_webhook" "build-webhook" {
  key = "build-webhook"
  event_types = ["uploaded", "deleted", "promoted"]
  criteria {
    any_build = false
    selected_builds = ["build-id"]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }
}

resource "artifactory_release_bundle_webhook" "release-bundle-webhook" {
  key = "release-bundle-webhook"
  event_types = ["created", "signed", "deleted"]
  criteria {
    any_release_bundle = false
    registered_release_bundle_names = ["release-bundle-name"]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }
}

resource "artifactory_distribution_webhook" "release-distribution-webhook" {
  key = "distribution-webhook"
  event_types = ["distribute_started", "distribute_completed", "distribute_aborted", "distribute_failed", "delete_started", "delete_completed", "delete_failed"]
  criteria {
    any_release_bundle = false
    registered_release_bundle_names = ["release-bundle-name"]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }
}

resource "artifactory_artifactory_release_bundle_webhook" "release-distribution-webhook" {
  key = "artifactory-release-bundle-webhook"
  event_types = ["received", "delete_started", "delete_completed", "delete_failed"]
  criteria {
    any_release_bundle = false
    registered_release_bundle_names = ["release-bundle-name"]
    include_patterns = ["foo/**"]
    exclude_patterns = ["bar/**"]
  }
  url = "http://tempurl.org/webhook"
  secret = "some-secret"
  proxy = "proxy-key"

  custom_http_headers = {
    header-1 = "value-1"
    header-2 = "value-2"
  }
}
