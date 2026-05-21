resource "artifactory_local_nix_repository" "nix-local" {
  key = "example-nix-local"
}

resource "artifactory_remote_nix_repository" "nix-remote" {
  key = "example-nix-remote"
  url = "https://cache.nixos.org"
}

resource "artifactory_virtual_nix_repository" "nix-virtual" {
  key = "example-nix-virtual"
  repositories = [
    artifactory_local_nix_repository.nix-local.key,
    artifactory_remote_nix_repository.nix-remote.key,
  ]
}
