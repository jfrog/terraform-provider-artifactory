# Artifactory Provider

The [Artifactory](https://jfrog.com/artifactory/) provider is used to interact with the resources supported by Artifactory. The provider needs to be configured with the proper credentials before it can be used.

Links to documentation for specific resources can be found in the table of contents to the left.

This provider requires access to Artifactory APIs, which are only available in the _licensed_ pro and enterprise editions. You can determine which license you have by accessing the following the URL `${host}/artifactory/api/system/licenses/`.

You can either access it via API, or web browser - it require admin level credentials.

```sh
curl -sL ${host}/artifactory/api/system/licenses/ | jq .
{
  "type" : "Enterprise Plus Trial",
  "validThrough" : "Jan 29, 2022",
  "licensedTo" : "JFrog Ltd"
}
```

## Terraform CLI version support

Current version support [Terraform Protocol v6](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol#protocol-version-6) which mean Terraform CLI version 1.0 and later. 

## Example Usage
```tf
# Required for Terraform 1.0 and up (https://www.terraform.io/upgrade-guides)
terraform {
  required_providers {
    artifactory = {
      source  = "jfrog/artifactory"
      version = "12.3.3"
    }
  }
}

# Configure the Artifactory provider
provider "artifactory" {
  url           = "${var.artifactory_url}/artifactory"
  access_token  = "${var.artifactory_access_token}"
  # Optional: supply a client certificate for mutual TLS
  # client_certificate_path     = pathexpand("~/.jfrog/client.pem")
  # client_certificate_key_path = pathexpand("~/.jfrog/client-key.pem")
}

# Create a new repository
resource "artifactory_local_pypi_repository" "pypi-libs" {
  key             = "pypi-libs"
  repo_layout_ref = "simple-default"
  description     = "A pypi repository for python packages"
}
```

## Authentication

The Artifactory provider supports two ways of authentication. The following methods are supported:
* Access Token
* API Key (deprecated)
* Terraform Cloud OIDC provider

### Access Token

Artifactory access tokens may be used via the Authorization header by providing the `access_token` attribute to the provider block. Getting this value from the environment is supported with `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` variables.

Usage:
```tf
# Configure the Artifactory provider
provider "artifactory" {
  url           = "artifactory.site.com/artifactory"
  access_token  = "abc...xy"
}
```

### API Key (deprecated)

!>An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform). In a future version (scheduled for end of Q3, 2023), the option to disable the usage/creation of API Keys will be available and set to disabled by default. Admins will be able to enable the usage/creation of API Keys. By end of Q4 2024, API Keys will be deprecated all together and the option to use them will no longer be available. See [JFrog API Key Deprecation Process](https://jfrog.com/help/r/jfrog-platform-administration-documentation/jfrog-api-key-deprecation-process).

~>If `access_token` attribute, `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` environment variable is set, the provider will ignore `api_key` attribute.

Artifactory API keys may be used via the `X-JFrog-Art-Api` header by providing the `api_key` attribute in the provider block.

Usage:
```tf
# Configure the Artifactory provider
provider "artifactory" {
  url     = "artifactory.site.com/artifactory"
  api_key = "abc...xy"
}
```

### Terraform Cloud OIDC Provider

If you are using this provider on Terraform Cloud and wish to use dynamic credentials instead of static access token for authentication with JFrog platform, you can leverage Terraform as the OIDC provider.

To setup dynamic credentials, follow these steps:
1. Configure Terraform Cloud as a generic OIDC provider
2. Set environment variable in your Terraform Workspace
3. Setup Terraform Cloud in your configuration

During the provider start up, if it finds env var `TFC_WORKLOAD_IDENTITY_TOKEN` it will use this token with your JFrog instance to exchange for a short-live access token. If that is successful, the provider will the access token for all subsequent API requests with the JFrog instance.

#### Configure Terraform Cloud as generic OIDC provider

Follow [confgure an OIDC integration](https://jfrog.com/help/r/jfrog-platform-administration-documentation/configure-an-oidc-integration). Enter a name for the provider, e.g. `terraform-cloud`. Use `https://app.terraform.io` for "Provider URL". Choose your own value for "Audience", e.g. `jfrog-terraform-cloud`.

Then [configure an identity mapping](https://jfrog.com/help/r/jfrog-platform-administration-documentation/configure-identity-mappings) with appropriate "Claims JSON" (e.g. `aud`, `sub` at minimum. See [Terraform Workload Identity - Configuring Trust with your Cloud Platform](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/workload-identity-tokens#configuring-trust-with-your-cloud-platform)), and select the "Token scope", "User", and "Service" as desired.

#### Set environment variable in your Terraform Workspace

In your workspace, add an environment variable `TFC_WORKLOAD_IDENTITY_AUDIENCE` with audience value (e.g. `jfrog-terraform-cloud`) from JFrog OIDC integration above. See [Manually Generating Workload Identity Tokens](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/manual-generation) for more details.

When a run starts on Terraform Cloud, it will create a workload identity token with the specified audience and assigns it to the environment variable `TFC_WORKLOAD_IDENTITY_TOKEN` for the provider to consume.

See [Generating Multiple Tokens](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/manual-generation#generating-multiple-tokens) on HCP Terraform for more details on using different tokens.

#### Setup Terraform Cloud in your configuration

Add `cloud` block to `terraform` block, and add `oidc_provider_name` attribute (from JFrog OIDC integration) to provider block:

```terraform
terraform {
  cloud {
    organization = "my-org"
    workspaces {
      name = "my-workspace"
    }
  }

  required_providers {
    artifactory = {
      source  = "jfrog/artifactory"
      version = "12.3.3"
    }
  }
}

provider "artifactory" {
  url = "https://myinstance.jfrog.io"
  oidc_provider_name = "terraform-cloud"
  tfc_credential_tag_name = "JFROG"
}
```

**Note:** Ensure `access_token` attribute and `JFROG_ACCESS_TOKEN` env var are not set

## Mutual TLS

Some Artifactory deployments require mutual TLS authentication. The provider can send a client certificate by either referencing local files or inlining PEM data.

To reference files:

```terraform
provider "artifactory" {
  url                         = "https://edge.example.com/artifactory"
  access_token                = var.artifactory_access_token
  client_certificate_path     = pathexpand("~/.jfrog/client-cert.pem")
  client_certificate_key_path = pathexpand("~/.jfrog/client-key.pem")
}
```

Use the same value for both path attributes if the PEM file contains the certificate and private key together. Both attributes must be provided when using the path-based configuration.

To inline PEM data (for example, when running on Terraform Cloud), supply both the certificate and matching private key:

```terraform
provider "artifactory" {
  url                    = "https://edge.example.com/artifactory"
  access_token           = var.artifactory_access_token
  client_certificate_pem = var.artifactory_client_certificate_pem
  client_private_key_pem = var.artifactory_client_private_key_pem
}
```

The following environment variables may also be used instead of configuration attributes:

- `JFROG_CLIENT_CERT_PATH` or `ARTIFACTORY_CLIENT_CERT_PATH`
- `JFROG_CLIENT_CERT_KEY_PATH` or `ARTIFACTORY_CLIENT_CERT_KEY_PATH`
- `JFROG_CLIENT_CERT_PEM` or `ARTIFACTORY_CLIENT_CERT_PEM`
- `JFROG_CLIENT_PRIVATE_KEY_PEM` or `ARTIFACTORY_CLIENT_PRIVATE_KEY_PEM`

All four variables participate in the same precedence rules as the provider attributes. File-based and inline options are mutually exclusive.

## Argument Reference

The following arguments are supported:

* `url` - (Optional) URL of Artifactory. This can also be sourced from the `JFROG_URL` or `ARTIFACTORY_URL` environment variable.
* `access_token` - (Optional) This can also be sourced from `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` environment variables.
* `api_key` - (Optional, deprecated) API key for api auth.
* `oidc_provider_name` - (Optional) OIDC provider name. See [Configure an OIDC Integration](https://jfrog.com/help/r/jfrog-platform-administration-documentation/configure-an-oidc-integration) for more details.
* `tfc_credential_tag_name` - (Optional) Terraform Cloud Workload Identity Token tag name. Use for generating multiple TFC workload identity tokens. When set, the provider will attempt to use env var with this tag name as suffix. **Note:** this is case sensitive, so if set to `JFROG`, then env var `TFC_WORKLOAD_IDENTITY_TOKEN_JFROG` is used instead of `TFC_WORKLOAD_IDENTITY_TOKEN`. See [Generating Multiple Tokens](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/manual-generation#generating-multiple-tokens) on HCP Terraform for more details.
* `client_certificate_path` - (Optional) Filesystem path to a PEM-encoded client certificate or certificate chain used for mutual TLS. Must be provided together with `client_certificate_key_path`. Can also be sourced from `JFROG_CLIENT_CERT_PATH` or `ARTIFACTORY_CLIENT_CERT_PATH`.
* `client_certificate_key_path` - (Optional) Filesystem path to the PEM-encoded private key that matches `client_certificate_path`. Can also be sourced from `JFROG_CLIENT_CERT_KEY_PATH` or `ARTIFACTORY_CLIENT_CERT_KEY_PATH`.
* `client_certificate_pem` - (Optional, Sensitive) Inline PEM-encoded client certificate or certificate chain used for mutual TLS. Must be provided together with `client_private_key_pem`. Can also be sourced from `JFROG_CLIENT_CERT_PEM` or `ARTIFACTORY_CLIENT_CERT_PEM`.
* `client_private_key_pem` - (Optional, Sensitive) Inline PEM-encoded private key that matches `client_certificate_pem`. Can also be sourced from `JFROG_CLIENT_PRIVATE_KEY_PEM` or `ARTIFACTORY_CLIENT_PRIVATE_KEY_PEM`.
