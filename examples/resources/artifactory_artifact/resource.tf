resource "artifactory_artifact" "my-artifact" {
  repository = "my-generic-local"
  path = "/my-path/my-file.zip"
  file_path = "/path/to/my-file.zip"
}