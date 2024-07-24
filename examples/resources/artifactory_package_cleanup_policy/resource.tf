resource "artifactory_package_cleanup_policy" "my-cleanup-policy" {
  key = "my-policy"
  description = "My package cleanup policy"
  cron_expression = "0 0 2 ? * MON-SAT *"
  duration_in_minutes = 60
  enabled = true
  skip_trashcan = false
  
  search_criteria = {
    package_types = ["docker"]
    repos = ["my-docker-local"]
    included_projects = ["myproj"]
    included_packages = ["**"]
    excluded_packages = ["com/jfrog/latest"]
    created_before_in_months = 1
    last_downloaded_before_in_months = 6
  }
}