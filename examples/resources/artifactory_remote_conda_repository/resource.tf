resource "artifactory_remote_conda_repository" "my-remote-conda" {
  key          = "my-remote-conda"
  url          = "https://repo.anaconda.com/pkgs/main"
  description  = "Remote Conda repository proxying Anaconda"
  curated      = true
  pass_through = false
}
