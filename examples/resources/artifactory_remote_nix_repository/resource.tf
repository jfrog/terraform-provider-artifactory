resource "artifactory_remote_nix_repository" "my-nix-remote" {
  key         = "my-nix-remote"
  url         = "https://cache.nixos.org"
  description = "Remote Nix binary cache"
}
