# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.25"
    }
  }
}

provider "artifactory" {
  //  supply ARTIFACTORY_USERNAME, _PASSWORD and _URL as env vars
}

resource "artifactory_remote_repository" "test-smart-remote" {
  allow_any_host_auth                   = false
  blacked_out                           = false
  block_mismatching_mime_types          = true
  bypass_head_requests                  = true
  description                           = "(local file cache)"
  download_context_path                 = "Download"
  enable_cookie_management              = false
  enable_token_authentication           = false
  fetch_jars_eagerly                    = false
  fetch_sources_eagerly                 = false
  force_nuget_authentication            = false
  handle_releases                       = true
  handle_snapshots                      = true
  hard_fail                             = false
  includes_pattern                      = "**/*"
  key                                   = "repo-remote"
  max_unique_snapshots                  = 0
  missed_cache_period_seconds           = 900
  notes                                 = "abcd note"
  offline                               = false
  package_type                          = "nuget"
  password                              = "***REMOVED***"
  property_sets                         = [
      "artifactory",
  ]
  remote_repo_checksum_policy_type      = "generate-if-absent"
  retrieval_cache_period_seconds        = 1800
  share_configuration                   = false
  socket_timeout_millis                 = 15000
  store_artifacts_locally               = true
  suppress_pom_consistency_checks       = true
  synchronize_properties                = false
  unused_artifacts_cleanup_period_hours = 168
  url                                   = "https://partnerenttest.jfrog.io/artifactory/api/nuget/nuget-local"
  username                              = "alexh@jfrog.com"
  v3_feed_url                           = "https://api.nuget.org/v3/index.json"
  xray_index                            = false

  // list_remote_folder_items = true
  content_synchronisation {
      enabled = true
      statistics_enabled = true
      properties_enabled = true
      source_origin_absence_detection_enabled = false
  }
}
