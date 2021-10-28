# Required for Terraform 0.13 and up (https://www.terraform.io/upgrade-guides/0-13.html)
terraform {
  required_providers {
    artifactory = {
      source  = "registry.terraform.io/jfrog/artifactory"
      version = "2.6.15"
    }
  }
}
resource "random_id" "randid" {
  byte_length = 16
}
resource "tls_private_key" "example" {
  algorithm   = "RSA"
  rsa_bits = 2048

}
resource "artifactory_keypair" "some-keypairRSA" {
  pair_name   = "some-keypairfoo"
  pair_type   = "RSA"
  alias       = "foo-aliasfoo"
  private_key = tls_private_key.example.private_key_pem
  public_key  = tls_private_key.example.public_key_pem
  depends_on = [tls_private_key.example,random_id.randid]
}
# currently PGP isn't supported
#resource "artifactory_keypair" "some-keypairPGP" {
#  pair_name   = "some-keypair${random_id.randid.id}"
#  pair_type   = "PGP"
#  alias       = "foo-alias${random_id.randid.id}"
#  private_key = file("samples/pgp.priv")
#  public_key  = file("samples/pgp.pub")
#  passphrase = "123456"
#}

resource "artifactory_local_alpine_repository" "terraform-local-test-repo-basic1896042683811651651" {
  key                 = "terraform-local-test-repo-basic1896042683811651651"
  primary_keypair_ref = artifactory_keypair.some-keypairRSA.pair_name
  depends_on = [artifactory_keypair.some-keypairRSA]
}