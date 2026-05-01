resource "artifactory_local_nix_repository" "my-nix-local" {
  key         = "my-nix-local"
  description = "Local Nix repository"
  notes       = "Internal repository"
}
