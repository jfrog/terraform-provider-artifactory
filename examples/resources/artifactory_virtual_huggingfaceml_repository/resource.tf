resource "artifactory_local_huggingfaceml_repository" "huggingfaceml-local" {
  key = "example-huggingfaceml-local"
}

resource "artifactory_remote_huggingfaceml_repository" "huggingfaceml-remote" {
  key = "example-huggingfaceml-remote"
}

resource "artifactory_virtual_huggingfaceml_repository" "huggingfaceml-virtual" {
  key = "example-huggingfaceml-virtual"
  repositories = [
    artifactory_local_huggingfaceml_repository.huggingfaceml-local.key,
    artifactory_remote_huggingfaceml_repository.huggingfaceml-remote.key,
  ]
}
