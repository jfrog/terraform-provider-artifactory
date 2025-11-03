resource "artifactory_keypair" "hex-keypair" {
  pair_name  = "hex-keypair"
  pair_type  = "RSA"
  alias      = "hex-alias"
  private_key = var.private_key
  public_key  = var.public_key
}

resource "artifactory_local_hex_repository" "local-hex-repo" {
  key                    = "local-hex-repo"
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
}

resource "artifactory_virtual_hex_repository" "my-hex-virtual" {
  key                    = "my-hex-virtual"
  repositories           = [artifactory_local_hex_repository.local-hex-repo.key]
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
}

