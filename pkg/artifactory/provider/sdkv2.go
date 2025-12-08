// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

// SdkV2 Artifactory provider that supports configuration via Access Token
// Supported resources are repos, users, groups, replications, and permissions
func SdkV2() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
				Description:      "Artifactory URL.",
			},
			"api_key": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				ValidateDiagFunc: validator.StringIsNotEmpty,
				Description:      "API key. If `access_token` attribute, `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` environment variable is set, the provider will ignore this attribute.",
				Deprecated: "An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform).\n" +
					"In a future version (scheduled for end of Q3, 2023), the option to disable the usage/creation of API Keys will be available and set to disabled by default. Admins will be able to enable the usage/creation of API Keys.\n" +
					"By end of Q4 2024, API Keys will be deprecated all together and the option to use them will no longer be available. See [JFrog API deprecation process](https://jfrog.com/help/r/jfrog-platform-administration-documentation/jfrog-api-key-deprecation-process) for more details.",
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "This is a access token that can be given to you by your admin under `User Management -> Access Tokens`. If not set, the 'api_key' attribute value will be used.",
			},
			"oidc_provider_name": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validator.StringIsNotEmpty,
				Description:      "OIDC provider name. See [Configure an OIDC Integration](https://jfrog.com/help/r/jfrog-platform-administration-documentation/configure-an-oidc-integration) for more details.",
			},
			"tfc_credential_tag_name": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validator.StringIsNotEmpty,
				Description:      "Terraform Cloud Workload Identity Token tag name. Use for generating multiple TFC workload identity tokens. When set, the provider will attempt to use env var with this tag name as suffix. **Note:** this is case sensitive, so if set to `JFROG`, then env var `TFC_WORKLOAD_IDENTITY_TOKEN_JFROG` is used instead of `TFC_WORKLOAD_IDENTITY_TOKEN`. See [Generating Multiple Tokens](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/manual-generation#generating-multiple-tokens) on HCP Terraform for more details.",
			},
		},

		ResourcesMap:   resourcesMap(),
		DataSourcesMap: datasourcesMap(),
	}

	p.ConfigureContextFunc = func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		ds := diag.Diagnostics{}
		meta, d := providerConfigure(ctx, data, p.TerraformVersion)
		if d != nil {
			ds = append(ds, d...)
		}
		return meta, ds
	}

	return p
}

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	// Check environment variables, first available OS variable will be assigned to the var
	url := util.CheckEnvVars([]string{"JFROG_URL", "ARTIFACTORY_URL"}, "")
	accessToken := util.CheckEnvVars([]string{"JFROG_ACCESS_TOKEN", "ARTIFACTORY_ACCESS_TOKEN"}, "")

	if v, ok := d.GetOk("url"); ok {
		url = v.(string)
	}
	if url == "" {
		return nil, diag.Errorf("missing URL Configuration")
	}

	restyClient, err := client.Build(url, productId)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if v, ok := d.GetOk("oidc_provider_name"); ok {
		oidcAccessToken, err := util.OIDCTokenExchange(ctx, restyClient, v.(string), d.Get("tfc_credential_tag_name").(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		if oidcAccessToken != "" {
			accessToken = oidcAccessToken
		}
	}

	if v, ok := d.GetOk("access_token"); ok && v != "" {
		accessToken = v.(string)
	}

	apiKey := d.Get("api_key").(string)

	restyClient, err = client.AddAuth(restyClient, apiKey, accessToken)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	bypassJFrogTLSVerification := os.Getenv("JFROG_BYPASS_TLS_VERIFICATION")
	if strings.ToLower(bypassJFrogTLSVerification) == "true" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		restyClient.SetTLSClientConfig(tlsConfig)
	}

	version, err := util.GetArtifactoryVersion(restyClient)
	if err != nil {
		return nil, diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  "Error getting Artifactory version",
			Detail:   fmt.Sprintf("The provider functionality might be affected by the absence of Artifactory version in the context. %v", err),
		}}
	}

	featureUsage := fmt.Sprintf("Terraform/%s", terraformVersion)
	go util.SendUsage(ctx, restyClient.R(), productId, featureUsage)

	return util.ProviderMetadata{
		Client:             restyClient,
		ArtifactoryVersion: version,
	}, nil
}
