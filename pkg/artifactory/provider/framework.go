package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v8/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-artifactory/v8/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/client"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

// Ensure the implementation satisfies the provider.Provider interface.
var _ provider.Provider = &ArtifactoryProvider{}

type ArtifactoryProvider struct {
	// Version is an example field that can be set with an actual provider
	// version on release, "dev" when the provider is built and ran locally,
	// and "test" when running acceptance testing.
	version string
}

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
			},
			"access_token": schema.StringAttribute{
				Description: "This is a access token that can be given to you by your admin under `Identity and Access`. If not set, the 'api_key' attribute value will be used.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_key": schema.StringAttribute{
				Description:        "API token. Projects functionality will not work with any auth method other than access tokens",
				DeprecationMessage: "An upcoming version will support the option to block the usage/creation of API Keys (for admins to set on their platform). In September 2022, the option to block the usage/creation of API Keys will be enabled by default, with the option for admins to change it back to enable API Keys. In January 2023, API Keys will be deprecated all together and the option to use them will no longer be available.",
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
	// Provider specific implementation.

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

	if config.Url.ValueString() != "" {
		url = config.Url.ValueString()
	}

	if accessToken == "" {
		resp.Diagnostics.AddError(
			"Missing  Access AccessToken Configuration",
			"While configuring the provider, the Access Token was not found in "+
				"the JFROG_ACCESS_TOKEN environment variable or provider "+
				"configuration block access_token attribute.",
		)
		return
	}

	if url == "" {
		resp.Diagnostics.AddError(
			"Missing URL Configuration",
			"While configuring the provider, the url was not found in "+
				"the JFROG_URL/ARTIFACTORY_URL environment variables or provider "+
				"configuration block url attribute.",
		)
		return
	}

	restyBase, err := client.Build(url, productId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Resty client",
			fmt.Sprintf("%v", err),
		)
	}
	restyBase, err = client.AddAuth(restyBase, "", accessToken)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding Auth to Resty client",
			fmt.Sprintf("%v", err),
		)
	}
	if config.CheckLicense.IsNull() || config.CheckLicense.ValueBool() == true {
		licenseErr := utilsdk.CheckArtifactoryLicense(restyBase, "Enterprise", "Commercial", "Edge")
		if licenseErr != nil {
			resp.Diagnostics.AddError(
				"Error getting Artifactory license",
				fmt.Sprintf("%v", licenseErr),
			)
			return
		}
	}

	version, err := utilsdk.GetArtifactoryVersion(restyBase)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Artifactory version",
			"The provider functionality might be affected by the absence of Artifactory version in the context.",
		)
		return
	}

	featureUsage := fmt.Sprintf("Terraform/%s", req.TerraformVersion)
	utilsdk.SendUsage(ctx, restyBase, "terraform-provider-artifactory/"+Version, featureUsage)

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
		security.NewPermissionTargetResource,
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
