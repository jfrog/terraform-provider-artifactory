resource "artifactory_package_cleanup_policy" "my-cleanup-policy" {
  key = "my-policy"
  description = "My package cleanup policy"
  cron_expression = "0 0 2 ? * MON-SAT *"
  duration_in_minutes = 60
  enabled = true
  skip_trashcan = false
  project_key = "myprojkey"
  
  search_criteria = {
    package_types = [
      "docker",
      "maven",
    ]
    repos = [
      "my-docker-local",
      "my-maven-local",
    ]
    excluded_repos = ["gradle-global"]
    include_all_projects = false
    included_projects = []
    included_packages = ["com/jfrog"]
    excluded_packages = ["com/jfrog/latest"]
    created_before_in_months = 1
    last_downloaded_before_in_months = 6
    keep_last_n_versions = 0
  }
}