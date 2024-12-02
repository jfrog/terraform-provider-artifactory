resource "artifactory_local_docker_v2_repository" "my-docker-local" {
  key             = "my-docker-local"
  tag_retention   = 3
  max_unique_tags = 5

  lifecycle {
    ignore_changes = ["project_key"]
  }
}

resource "project" "myproj" {
  key = "myproj"
  display_name = "Test Project"
  description  = "Test Project"
  admin_privileges {
    manage_members   = true
    manage_resources = true
    index_resources  = true
  }
  max_storage_in_gibibytes   = 10
  block_deployments_on_limit = false
  email_notification         = true
}

resource "project_repository" "myproj-my-docker-local" {
  project_key = project.myproj.key
  key = artifactory_local_docker_v2_repository.my-docker-local.key
}

resource "artifactory_archive_policy" "my-archive-policy" {
  key = "my-archive-policy"
  description = "My archive policy"
  cron_expression = "0 0 2 ? * MON-SAT *"
  duration_in_minutes = 60
  enabled = true
  skip_trashcan = false
  project_key = project_repository.myproj-my-docker-local.project_key
  
  search_criteria = {
    package_types = [
      "docker",
    ]
    repos = [
      project_repository.myproj-my-docker-local.key,
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