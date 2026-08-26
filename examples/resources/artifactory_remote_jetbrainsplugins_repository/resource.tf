resource "artifactory_remote_jetbrainsplugins_repository" "my-jetbrainsplugins-remote" {
  key                            = "my-jetbrainsplugins-remote"
  url                            = "https://plugins.jetbrains.com"
  description                    = "Remote JetBrains Plugins repository proxying the JetBrains Marketplace"
  notes                          = "Internal repository"
  retrieval_cache_period_seconds = 7200
  bypass_head_requests           = false
  list_remote_folder_items       = false
}
