package repository

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
)

// BaseRepositoryDataSource provides common functionality for repository data sources
type BaseRepositoryDataSource struct {
	ProviderData util.ProviderMetadata
}

// BaseRepositoryDataSourceModel contains common fields for all repository data sources
type BaseRepositoryDataSourceModel struct {
	Key                    types.String `tfsdk:"key"`
	Description            types.String `tfsdk:"description"`
	Notes                  types.String `tfsdk:"notes"`
	ProjectKey             types.String `tfsdk:"project_key"`
	ProjectEnvironments    types.Set    `tfsdk:"project_environments"`
	RepoLayoutRef          types.String `tfsdk:"repo_layout_ref"`
	BlackedOut             types.Bool   `tfsdk:"blacked_out"`
	XrayIndex              types.Bool   `tfsdk:"xray_index"`
	PropertySets           types.Set    `tfsdk:"property_sets"`
	ArchiveBrowsingEnabled types.Bool   `tfsdk:"archive_browsing_enabled"`
	DownloadDirect         types.Bool   `tfsdk:"download_direct"`
	PriorityResolution     types.Bool   `tfsdk:"priority_resolution"`
	CDNRedirect            types.Bool   `tfsdk:"cdn_redirect"`
}

// BaseRepositoryAPIModel contains common fields for all repository API models
type BaseRepositoryAPIModel struct {
	Key                    string   `json:"key"`
	Description            string   `json:"description"`
	Notes                  string   `json:"notes"`
	ProjectKey             string   `json:"projectKey"`
	ProjectEnvironments    []string `json:"projectEnvironments"`
	RepoLayoutRef          string   `json:"repoLayoutRef"`
	BlackedOut             bool     `json:"blackedOut"`
	XrayIndex              bool     `json:"xrayIndex"`
	PropertySets           []string `json:"propertySets"`
	ArchiveBrowsingEnabled bool     `json:"archiveBrowsingEnabled"`
	DownloadDirect         bool     `json:"downloadDirect"`
	PriorityResolution     bool     `json:"priorityResolution"`
	CDNRedirect            bool     `json:"cdnRedirect"`
}

// BaseSchemaAttributes returns the base schema attributes for all repository data sources
func BaseSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"key": schema.StringAttribute{
			MarkdownDescription: "The identity key of the repository",
			Required:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Public description",
			Computed:            true,
		},
		"notes": schema.StringAttribute{
			MarkdownDescription: "Internal description",
			Computed:            true,
		},
		"project_key": schema.StringAttribute{
			MarkdownDescription: "Project key for assigning this repository to. Must be 2 - 20 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
			Computed:            true,
		},
		"project_environments": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "Project environment for assigning this repository to. Allow values: \"DEV\", \"PROD\", or one of custom value. Before Artifactory 7.53.1, up to 2 values (\"DEV\" and \"PROD\") are allowed. From 7.53.1 onward, only one value is allowed. The attribute should only be used if the repository is already assigned to the existing project. If not, the attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create state drift during the update.",
			Computed:            true,
		},
		"repo_layout_ref": schema.StringAttribute{
			MarkdownDescription: "Repository layout key for the local repository",
			Computed:            true,
		},
		"blacked_out": schema.BoolAttribute{
			MarkdownDescription: "When set, the repository will use the deprecated trivial layout in which the URL path to the file is the same as the repository path. This option is useful for maintaining compatibility with existing tools while preserving the new layout structure. Default value is `false`.",
			Computed:            true,
		},
		"xray_index": schema.BoolAttribute{
			MarkdownDescription: "Enable Indexing In Xray. Repository will be indexed with the default retention period. You will not be able to change the retention period per repository. This setting applies to all repository types except for the following: Local Repository: Alpine, Cargo, CocoaPods, Composer, Conan, Conda, CRAN, Debian, Go, Helm, HPM, Ivy, Maven, NPM, NuGet, OPKG, P2, Pub, Puppet, R, RPM, SBT, Swift, Vagrant, YUM, and Zypper. Remote Repository: Alpine, Cargo, CocoaPods, Composer, Conan, Conda, CRAN, Debian, Go, Helm, HPM, Ivy, Maven, NPM, NuGet, OPKG, P2, Pub, Puppet, R, RPM, SBT, Swift, Vagrant, YUM, and Zypper. Virtual Repository: Alpine, Cargo, CocoaPods, Composer, Conan, Conda, CRAN, Debian, Go, Helm, HPM, Ivy, Maven, NPM, NuGet, OPKG, P2, Pub, Puppet, R, RPM, SBT, Swift, Vagrant, YUM, and Zypper.",
			Computed:            true,
		},
		"property_sets": schema.SetAttribute{
			ElementType:         types.StringType,
			MarkdownDescription: "List of property set names",
			Computed:            true,
		},
		"archive_browsing_enabled": schema.BoolAttribute{
			MarkdownDescription: "When set, you may view content such as HTML or Javadoc files directly from Artifactory. This may not be safe and therefore requires strict content moderation to prevent malicious users from uploading content that may compromise security (e.g., cross-site scripting attacks).",
			Computed:            true,
		},
		"download_direct": schema.BoolAttribute{
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the file directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.",
			Computed:            true,
		},
		"priority_resolution": schema.BoolAttribute{
			MarkdownDescription: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
			Computed:            true,
		},
		"cdn_redirect": schema.BoolAttribute{
			MarkdownDescription: "When set, Artifactory will return an error to the client that causes the build to fail if there is a failure to communicate with this repository.",
			Computed:            true,
		},
	}
}

// CommonConfigure provides common configuration logic for repository data sources
func CommonConfigure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse, ds interface{}) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// Type assertion to set ProviderData
	if baseDS, ok := ds.(interface{ SetProviderData(util.ProviderMetadata) }); ok {
		baseDS.SetProviderData(req.ProviderData.(util.ProviderMetadata))
	}
}

// CommonRead provides common read logic for repository data sources
func CommonRead(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse, ds interface{}, key string) {
	// Get the data source that implements the required methods
	readableDS, ok := ds.(interface {
		GetData() interface{}
		SetData(interface{})
		FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics
		GetProviderData() util.ProviderMetadata
	})

	if !ok {
		resp.Diagnostics.AddError(
			"Invalid Data Source",
			"The data source does not implement the required interface for common read operations.",
		)
		return
	}

	// Get the data model
	data := readableDS.GetData()

	// Parse the configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make the API call
	var apiModel interface{}
	response, err := readableDS.GetProviderData().Client.R().
		SetResult(&apiModel).
		Get("artifactory/api/repositories/" + key)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetching the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	if response.IsError() {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetching the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+response.String(),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	resp.Diagnostics.Append(readableDS.FromAPIModel(ctx, apiModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

// CommonFromAPIModel provides common conversion logic from API model to Terraform model
func CommonFromAPIModel(ctx context.Context, baseModel *BaseRepositoryDataSourceModel, apiModel BaseRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	baseModel.Key = types.StringValue(apiModel.Key)
	baseModel.Description = types.StringValue(apiModel.Description)
	baseModel.Notes = types.StringValue(apiModel.Notes)
	baseModel.ProjectKey = types.StringValue(apiModel.ProjectKey)
	baseModel.RepoLayoutRef = types.StringValue(apiModel.RepoLayoutRef)
	baseModel.BlackedOut = types.BoolValue(apiModel.BlackedOut)
	baseModel.XrayIndex = types.BoolValue(apiModel.XrayIndex)
	baseModel.ArchiveBrowsingEnabled = types.BoolValue(apiModel.ArchiveBrowsingEnabled)
	baseModel.DownloadDirect = types.BoolValue(apiModel.DownloadDirect)
	baseModel.PriorityResolution = types.BoolValue(apiModel.PriorityResolution)
	baseModel.CDNRedirect = types.BoolValue(apiModel.CDNRedirect)

	// Convert property sets
	if apiModel.PropertySets != nil {
		propertySets, diags := types.SetValueFrom(ctx, types.StringType, apiModel.PropertySets)
		if diags.HasError() {
			return diags
		}
		baseModel.PropertySets = propertySets
	} else {
		baseModel.PropertySets = types.SetNull(types.StringType)
	}

	// Convert project environments
	if apiModel.ProjectEnvironments != nil {
		projectEnvironments, diags := types.SetValueFrom(ctx, types.StringType, apiModel.ProjectEnvironments)
		if diags.HasError() {
			return diags
		}
		baseModel.ProjectEnvironments = projectEnvironments
	} else {
		baseModel.ProjectEnvironments = types.SetNull(types.StringType)
	}

	return diags
}
