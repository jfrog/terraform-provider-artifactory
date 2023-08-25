package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

var Version = "7.0.0" // needs to be exported so make file can update this
var productId = "terraform-provider-artifactory/" + Version

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
				ConflictsWith:    []string{"access_token"},
				ValidateDiagFunc: validator.StringIsNotEmpty,
				Description:      "API token. Projects functionality will not work with any auth method other than access tokens",
				Deprecated: "An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform).\n" +
					"In a future version (scheduled for end of Q3, 2023), the option to disable the usage/creation of API Keys will be available and set to disabled by default. Admins will be able to enable the usage/creation of API Keys.\n" +
					"By end of Q1 2024, API Keys will be deprecated all together and the option to use them will no longer be available.",
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "This is a access token that can be given to you by your admin under `Identity and Access`. If not set, the 'api_key' attribute value will be used.",
			},
			"check_license": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Toggle for pre-flight checking of Artifactory Pro and Enterprise license. Default to `true`.",
			},
		},

		ResourcesMap:   resourcesMap(),
		DataSourcesMap: datasourcesMap(),
	}

	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		tflog.Debug(ctx, "ConfigureContextFunc")
		tflog.Info(ctx, fmt.Sprintf("Provider version: %s", Version))

		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(ctx, d, terraformVersion)
	}

	return p
}

// Creates the client for artifactory, will prefer token auth over basic auth if both set
func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {
	tflog.Debug(ctx, "providerConfigure")

	// Check environment variables, first available OS variable will be assigned to the var
	url := CheckEnvVars([]string{"JFROG_URL", "ARTIFACTORY_URL"}, "http://localhost:8082")
	accessToken := CheckEnvVars([]string{"JFROG_ACCESS_TOKEN", "ARTIFACTORY_ACCESS_TOKEN"}, "")

	if d.Get("url") != "" {
		url = d.Get("url").(string)
	}
	if d.Get("access_token") != "" {
		accessToken = d.Get("access_token").(string)
	}
	apiKey := d.Get("api_key").(string)

	restyBase, err := client.Build(url, productId)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	restyBase, err = client.AddAuth(restyBase, apiKey, accessToken)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	// Due to migration from SDK v2 to plugin framework, we have to remove defaults from the provider configuration.
	// https://discuss.hashicorp.com/t/muxing-upgraded-tfsdk-and-framework-provider-with-default-provider-configuration/43945
	checkLicense := true
	v, checkLicenseBoolSet := d.GetOkExists("check_license")
	if checkLicenseBoolSet {
		checkLicense = v.(bool)
	}
	if checkLicense {
		licenseErr := utilsdk.CheckArtifactoryLicense(restyBase, "Enterprise", "Commercial", "Edge")
		if licenseErr != nil {
			return nil, licenseErr
		}
	}

	version, err := utilsdk.GetArtifactoryVersion(restyBase)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	featureUsage := fmt.Sprintf("Terraform/%s", terraformVersion)
	utilsdk.SendUsage(ctx, restyBase, productId, featureUsage)

	return utilsdk.ProvderMetadata{
		Client:             restyBase,
		ArtifactoryVersion: version,
	}, nil
}
