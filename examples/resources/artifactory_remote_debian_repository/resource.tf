resource "artifactory_remote_debian_repository" "my-remote-debian" {
  key          = "my-remote-debian"
  url          = "http://archive.ubuntu.com/ubuntu/"
  description  = "Remote Debian repository proxying Ubuntu archive"
  curated      = true
  pass_through = false
}
