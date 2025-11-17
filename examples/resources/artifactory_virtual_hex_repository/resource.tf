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

resource "artifactory_local_hex_repository" "local-hex" {
  key                     = "local-hex"
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
  depends_on              = [artifactory_keypair.hex-keypair]
}

resource "artifactory_remote_hex_repository" "remote-hex" {
  key                     = "remote-hex"
  url                     = "https://repo.hex.pm"
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
  public_key              = file("samples/rsa.pub")
  depends_on              = [artifactory_keypair.hex-keypair]
}

resource "artifactory_virtual_hex_repository" "my-hex-virtual" {
  key                     = "my-hex-virtual"
  hex_primary_keypair_ref = artifactory_keypair.hex-keypair.pair_name
  repositories           = [
    artifactory_local_hex_repository.local-hex.key,
    artifactory_remote_hex_repository.remote-hex.key
  ]
  description             = "Virtual Hex repository aggregating local and remote"
  notes                   = "Internal repository"
  depends_on              = [
    artifactory_keypair.hex-keypair,
    artifactory_local_hex_repository.local-hex,
    artifactory_remote_hex_repository.remote-hex
  ]
}

