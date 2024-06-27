resource "artifactory_vault_configuration" "my-vault-config-app-role" {
  name = "my-vault-config-app-role"
  config = {
    url = "http://127.0.0.1:8200"
    auth = {
      type      = "AppRole"
      role_id   = "1b62ff05..."
      secret_id = "acbd6657..."
    }

    mounts = [
      {
        path = "secret"
        type = "KV2"
      }
    ]
  }
}

resource "artifactory_vault_configuration" "my-vault-config-cert" {
  name = "my-vault-config-cert"
  config = {
    url = "http://127.0.0.1:8200"
    auth = {
      type            = "Certificate"
      certificate     = file("samples/public.pem")
      certificate_key = file("samples/private.pem")
    }

    mounts = [
      {
        path = "secret"
        type = "KV2"
      }
    ]
  }
}