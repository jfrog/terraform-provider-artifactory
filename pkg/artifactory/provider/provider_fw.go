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
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
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
	CheckLicense types.Bool   `tfsdk:"check_license"`
}

// Metadata satisfies the provider.Provider interface for ArtifactoryProvider
func (p *ArtifactoryProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "terraform-provider-artifactory"
	resp.Version = "7.0.0" // Will be to overwritten in a build process
	resp.Version = p.version
}

// Schema satisfies the provider.Provider interface for ArtifactoryProvider.
func (p *ArtifactoryProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Artifactory URL.",
				Optional:            true,
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "This is a access token that can be given to you by your admin under `Identity and Access`. If not set, the 'api_key' attribute value will be used.",
				Optional:            true,
				Sensitive:           true,
			},
			"check_license": schema.BoolAttribute{
				MarkdownDescription: "Toggle for pre-flight checking of Artifactory Pro and Enterprise license. Default to `true`.",
				Optional:            true,
			},
		},
	}
}

func (p *ArtifactoryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Provider specific implementation.

	// Check environment variables first
	url := os.Getenv("JFROG_URL")
	accessToken := os.Getenv("JFROG_ACCESS_TOKEN")

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
			"While configuring the provider, the API token was not found in "+
				"the JFROG_ACCESS_TOKEN environment variable or provider "+
				"configuration block access_token attribute.",
		)
		// Not returning early allows the logic to collect all errors.
	}

	if url == "" {
		resp.Diagnostics.AddError(
			"Missing URL Configuration",
			"While configuring the provider, the endpoint was not found in "+
				"the JFROG_URL environment variable or provider "+
				"configuration block url attribute.",
		)
		// Not returning early allows the logic to collect all errors.
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
	if config.CheckLicense.IsNull() {
		licenseErr := util.CheckArtifactoryLicense(restyBase, "Enterprise", "Commercial", "Edge")
		if licenseErr != nil {
			resp.Diagnostics.AddError(
				"Error getting Artifactory license",
				fmt.Sprintf("%v", err),
			)
			return
		}
	}

	version, err := util.GetArtifactoryVersion(restyBase)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Artifactory version",
			"The provider functionality might be affected by the absence of Artifactory version in the context.",
		)
	}

	featureUsage := fmt.Sprintf("Terraform/%s", req.TerraformVersion)
	util.SendUsage(ctx, restyBase, "terraform-provider-artifactory/"+p.version, featureUsage)

	resp.DataSourceData = util.ProvderMetadata{
		Client:             restyBase,
		ArtifactoryVersion: version,
	}

	resp.ResourceData = util.ProvderMetadata{
		Client:             restyBase,
		ArtifactoryVersion: version,
	}

}

// Resources satisfies the provider.Provider interface for ArtifactoryProvider.
func (p *ArtifactoryProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		user.NewArtifactoryUserResource,
		user.NewArtifactoryManagedUserResource,
		user.NewArtifactoryAnonymousUserResource,
	}
}

// DataSources satisfies the provider.Provider interface for ArtifactoryProvider.
func (p *ArtifactoryProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Provider specific implementation
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ArtifactoryProvider{
			version: version,
		}
	}
}
