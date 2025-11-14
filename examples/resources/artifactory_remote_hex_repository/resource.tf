resource "artifactory_keypair" "hex-keypair" {
  pair_name   = "hex-keypair"
  pair_type   = "RSA"
  alias       = "hex-alias"
  private_key = file("samples/rsa.priv")
  public_key  = file("samples/rsa.pub")
  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_remote_hex_repository" "my-hex-remote" {
  key                     = "my-hex-remote"
  url                     = "https://repo.hex.pm"
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
  public_key              = file("samples/rsa.pub")
  description             = "Remote Hex repository for Elixir packages"
  notes                   = "Internal repository"
  depends_on              = [artifactory_keypair.hex-keypair]
}

