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

resource "artifactory_release_bundle_v2_cleanup_policy" "my-resource-bundle-v2-cleanup-policy" {
  key                 = "my-release-bundle-v2-policy-key"
  description         = "Cleanup policy description"
  cron_expression     = "0 0 2 * * ?"
  duration_in_minutes = 60
  enabled             = true
  search_criteria = {
    include_all_projects = true
    included_projects    = []
    release_bundles = [
      {
        name        = "my-release-bundle-v2-rb"
        project_key = ""
      }
    ]
    exclude_promoted_environments = [
      "**"
    ]
  }
}