resource "artifactory_keypair" "hex-keypair" {
  pair_name  = "hex-keypair"
  pair_type  = "RSA"
  alias      = "hex-alias"
  private_key = var.private_key
  public_key  = var.public_key
}

resource "artifactory_remote_hex_repository" "my-hex-remote" {
  key                    = "my-hex-remote"
  url                    = "https://hex.pm/"
  public_key_ref         = file("${path.module}/hex_public_key")
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
}

