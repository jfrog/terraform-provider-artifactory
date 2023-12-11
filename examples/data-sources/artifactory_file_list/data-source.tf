data "artifactory_file_list" "my-repo-file-list" {
  repository_key = "my-generic-local"
  folder_path    = "path/to/artifact"
}