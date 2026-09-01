resource "artifactory_keypair" "terraform-signing-key" {
  pair_name   = "terraform-signing-key"
  pair_type   = "GPG"
  alias       = "terraform-provider-signing"
  private_key = file("samples/gpg.priv")
  public_key  = file("samples/gpg.pub")

  lifecycle {
    ignore_changes = [
      private_key,
      passphrase,
    ]
  }
}

resource "artifactory_federated_terraform_provider_repository" "terraform-federated-test-terraform-provider-repo" {
  key                 = "terraform-federated-test-terraform-provider-repo"
  primary_keypair_ref = artifactory_keypair.terraform-signing-key.pair_name

  member {
    url     = "http://tempurl.org/artifactory/terraform-federated-test-terraform-provider-repo"
    enabled = true
  }

  member {
    url     = "http://tempurl2.org/artifactory/terraform-federated-test-terraform-provider-repo-2"
    enabled = true
  }

  depends_on = [artifactory_keypair.terraform-signing-key]
}
