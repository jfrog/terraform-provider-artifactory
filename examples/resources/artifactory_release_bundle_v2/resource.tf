resource "artifactory_release_bundle_v2" "my-release-bundle-v2-aql" {
  name                            = "my-release-bundle-v2-aql"
  version                         = "1.0.0"
  keypair_name                    = "my-keypair-name"
  project_key                     = "myproj-key"
  skip_docker_manifest_resolution = true
  source_type                     = "aql"

  source = {
    aql = "items.find({\"repo\": {\"$match\": \"my-generic-*\"}})"
  }
}

resource "artifactory_release_bundle_v2" "my-release-bundle-v2-artifacts" {
  name                            = "my-release-bundle-v2-artifacts"
  version                         = "1.0.0"
  keypair_name                    = "my-keypair-name"
  skip_docker_manifest_resolution = true
  source_type                     = "artifacts"

  source = {
    artifacts = [{
      path   = "commons-qa-maven-local/org/apache/tomcat/commons/1.0.0/commons-1.0.0.jar"
      sha256 = "0d2053f76605e0734f5251a78c5dade5ee81b0f3730b3f603aedb90bc58033fb"
    }]
  }
}

resource "artifactory_release_bundle_v2" "my-release-bundle-v2-builds" {
  name                            = "my-release-bundle-v2-builds"
  version                         = "1.0.0"
  keypair_name                    = "my-keypair-name"
  skip_docker_manifest_resolution = true
  source_type                     = "builds"

  source = {
    builds = [{
      name   = "my-build-info-name"
      number = "1.0"
    }]
  }
}

resource "artifactory_release_bundle_v2" "my-release-bundle-v2-rb" {
  name                            = "my-release-bundle-v2-rb"
  version                         = "2.0.0"
  keypair_name                    = "my-keypair-name"
  skip_docker_manifest_resolution = true
  source_type                     = "release_bundles"

  source = {
    release_bundles = [{
      name    = "my-rb-name"
      version = "1.0.0"
    }]
  }
}