package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

var Version = "7.0.0" // needs to be exported so make file can update this
var productId = "terraform-provider-artifactory/" + Version

// Provider Artifactory provider that supports configuration via Access Token
// Supported resources are repos, users, groups, replications, and permissions
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": {
				Type:             schema.TypeString,
				Optional:         true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ARTIFACTORY_URL", "JFROG_URL"}, "http://localhost:8082"),
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
			},
			"api_key": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				DefaultFunc:      schema.EnvDefaultFunc("ARTIFACTORY_API_KEY", nil),
				ConflictsWith:    []string{"access_token"},
				ValidateDiagFunc: validator.StringIsNotEmpty,
				Description:      "API token. Projects functionality will not work with any auth method other than access tokens",
				Deprecated: "An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform).\n" +
					"In September 2022, the option to block the usage/creation of API Keys will be enabled by default, with the option for admins to change it back to enable API Keys.\n" +
					"In January 2023, API Keys will be deprecated all together and the option to use them will no longer be available.",
			},
			"access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARTIFACTORY_ACCESS_TOKEN", "JFROG_ACCESS_TOKEN"}, nil),
				Description: "This is a access token that can be given to you by your admin under `Identity and Access`. If not set, the 'api_key' attribute value will be used.",
			},
			"check_license": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
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

	URL := d.Get("url")

	restyBase, err := client.Build(URL.(string), productId)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	apiKey := d.Get("api_key").(string)
	accessToken := d.Get("access_token").(string)

	restyBase, err = client.AddAuth(restyBase, apiKey, accessToken)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	checkLicense := d.Get("check_license").(bool)
	if checkLicense {
		licenseErr := util.CheckArtifactoryLicense(restyBase, "Enterprise", "Commercial", "Edge")
		if licenseErr != nil {
			return nil, licenseErr
		}
	}

	version, err := util.GetArtifactoryVersion(restyBase)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	featureUsage := fmt.Sprintf("Terraform/%s", terraformVersion)
	util.SendUsage(ctx, restyBase, productId, featureUsage)

	return util.ProvderMetadata{
		Client:             restyBase,
		ArtifactoryVersion: version,
	}, nil
}
