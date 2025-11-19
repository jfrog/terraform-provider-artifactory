provider "artifactory" {
  url          = "https://myinstance.jfrog.io"
  access_token = "my-access-token"
  # Optional: configure mutual TLS if required by your Artifactory instance
  # client_certificate_path     = pathexpand("~/.jfrog/client-cert.pem")
  # client_certificate_key_path = pathexpand("~/.jfrog/client-key.pem")
}
