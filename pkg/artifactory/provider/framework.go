package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/client"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
)

// Ensure the implementation satisfies the provider.Provider interface.
var _ provider.Provider = &ArtifactoryProvider{}

type ArtifactoryProvider struct{}

// ArtifactoryProviderModel describes the provider data model.
type ArtifactoryProviderModel struct {
	Url          types.String `tfsdk:"url"`
	AccessToken  types.String `tfsdk:"access_token"`
	ApiKey       types.String `tfsdk:"api_key"`
	CheckLicense types.Bool   `tfsdk:"check_license"`
}

// Metadata satisfies the provider.Provider interface for ArtifactoryProvider
func (p *ArtifactoryProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "terraform-provider-artifactory"
	resp.Version = Version
}

// Schema satisfies the provider.Provider interface for ArtifactoryProvider.
func (p *ArtifactoryProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "Artifactory URL.",
				Optional:    true,
				Validators: []validator.String{
					validatorfw_string.IsURLHttpOrHttps(),
				},
			},
			"access_token": schema.StringAttribute{
				Description: "This is a access token that can be given to you by your admin under `User Management -> Access Tokens`. If not set, the 'api_key' attribute value will be used.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"api_key": schema.StringAttribute{
				Description:        "API key. If `access_token` attribute, `JFROG_ACCESS_TOKEN` or `ARTIFACTORY_ACCESS_TOKEN` environment variable is set, the provider will ignore this attribute.",
				DeprecationMessage: "An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform). In a future version (scheduled for end of Q3, 2023), the option to disable the usage/creation of API Keys will be available and set to disabled by default. Admins will be able to enable the usage/creation of API Keys. By end of Q1 2024, API Keys will be deprecated all together and the option to use them will no longer be available.",
				Optional:           true,
				Sensitive:          true,
			},
			"check_license": schema.BoolAttribute{
				Description: "Toggle for pre-flight checking of Artifactory Pro and Enterprise license. Default to `true`.",
				Optional:    true,
			},
		},
	}
}

func (p *ArtifactoryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// check if Terraform version is >=1.0.0, i.e. support protocol v6
	supportProtocolV6, err := utilsdk.CheckVersion(req.TerraformVersion, "1.0.0")
	if err != nil {
		resp.Diagnostics.Append(diag.NewWarningDiagnostic("failed to check Terraform version", err.Error()))
	}

	if !supportProtocolV6 {
		resp.Diagnostics.Append(diag.NewWarningDiagnostic(
			"Terraform CLI version deprecation",
			"Terraform version older than 1.0 will no longer be supported in Q1 2024. Please upgrade to latest Terraform CLI.",
		))
	}

	// Check environment variables, first available OS variable will be assigned to the var
	url := CheckEnvVars([]string{"JFROG_URL", "ARTIFACTORY_URL"}, "")
	accessToken := CheckEnvVars([]string{"JFROG_ACCESS_TOKEN", "ARTIFACTORY_ACCESS_TOKEN"}, "")

	var config ArtifactoryProviderModel

	// Read configuration data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check configuration data, which should take precedence over
	// environment variable data, if found.
	if config.AccessToken.ValueString() != "" {
		accessToken = config.AccessToken.ValueString()
	}

	if accessToken == "" {
		resp.Diagnostics.AddError(
			"Missing JFrog Access Token",
			"While configuring the provider, the Access Token was not found in "+
				"the JFROG_ACCESS_TOKEN/ARTIFACTORY_ACCESS_TOKEN environment variable or provider "+
				"configuration block access_token attribute.",
		)
		return
	}

	if config.Url.ValueString() != "" {
		url = config.Url.ValueString()
	}

	if url == "" {
		resp.Diagnostics.AddError(
			"Missing URL Configuration",
			"While configuring the provider, the url was not found in "+
				"the JFROG_URL/ARTIFACTORY_URL environment variable or provider "+
				"configuration block url attribute.",
		)
		return
	}

	restyBase, err := client.Build(url, productId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resty client",
			err.Error(),
		)
	}

	restyBase, err = client.AddAuth(restyBase, config.ApiKey.ValueString(), accessToken)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Auth to Resty client",
			err.Error(),
		)
	}

	if config.CheckLicense.IsNull() || config.CheckLicense.ValueBool() {
		if licenseErr := utilsdk.CheckArtifactoryLicense(restyBase, "Enterprise", "Commercial", "Edge"); licenseErr != nil {
			resp.Diagnostics.AddError(
				"Error getting Artifactory license",
				fmt.Sprintf("%v", licenseErr),
			)
			return
		}
	}

	version, err := utilsdk.GetArtifactoryVersion(restyBase)
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Error getting Artifactory version",
			fmt.Sprintf("The provider functionality might be affected by the absence of Artifactory version in the context. %v", err),
		)
		return
	}

	featureUsage := fmt.Sprintf("Terraform/%s", req.TerraformVersion)
	utilsdk.SendUsage(ctx, restyBase, productId, featureUsage)

	resp.DataSourceData = utilsdk.ProvderMetadata{
		Client:             restyBase,
		ArtifactoryVersion: version,
	}

	resp.ResourceData = utilsdk.ProvderMetadata{
		Client:             restyBase,
		ArtifactoryVersion: version,
	}
}

// Resources satisfies the provider.Provider interface for ArtifactoryProvider.
func (p *ArtifactoryProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		user.NewUserResource,
		user.NewManagedUserResource,
		user.NewAnonymousUserResource,
		security.NewGroupResource,
		security.NewScopedTokenResource,
		security.NewGlobalEnvironmentResource,
		security.NewDistributionPublicKeyResource,
		security.NewCertificateResource,
		security.NewKeyPairResource,
		configuration.NewLdapSettingResource,
		configuration.NewLdapGroupSettingResource,
		configuration.NewBackupResource,
		configuration.NewMailServerResource,
	}
}

// DataSources satisfies the provider.Provider interface for ArtifactoryProvider.
func (p *ArtifactoryProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Provider specific implementation
	}
}

func Framework() func() provider.Provider {
	return func() provider.Provider {
		return &ArtifactoryProvider{}
	}
}

func CheckEnvVars(vars []string, dv string) string {
	for _, k := range vars {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return dv
}
