resource "artifactory_remote_bazel_repository" "my-bazelmodules-remote" {
  key         = "my-bazelmodules-remote"
  url         = "https://bcr.bazel.build/"
  description = "Remote Bazel Modules repository proxying the Bazel Central Registry"
  notes       = "Internal repository"
}
