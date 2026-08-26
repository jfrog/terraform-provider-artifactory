resource "artifactory_remote_aieditorextensions_repository" "my-aieditorextensions-remote" {
  key                            = "my-aieditorextensions-remote"
  url                            = "https://marketplace.visualstudio.com/_apis/public/gallery"
  description                    = "Remote AI-Editor Extensions repository proxying the VS Code marketplace gallery"
  notes                          = "Internal repository"
  external_dependencies_enabled  = true
  external_dependencies_patterns = ["**/**vsassets.io/**"]
  curated                        = false
  pass_through                   = false
}
