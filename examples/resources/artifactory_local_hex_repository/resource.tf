resource "artifactory_keypair" "hex-keypair" {
  pair_name  = "hex-keypair"
  pair_type  = "RSA"
  alias      = "hex-alias"
  private_key = var.private_key
  public_key  = var.public_key
}

resource "artifactory_local_hex_repository" "my-hex-local" {
  key                    = "my-hex-local"
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
}

