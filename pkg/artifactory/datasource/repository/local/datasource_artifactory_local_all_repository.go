package local

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/repository"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

var _ datasource.DataSource = &AllLocalDataSource{}

func NewAllLocalDataSource() datasource.DataSource {
	return &AllLocalDataSource{}
}

type AllLocalDataSource struct {
	ProviderData utilsdk.ProvderMetadata
}

type LocalDataSourceModel struct {
	repository.DataSourceModel
	DebianTrivialLayout             types.Bool   `tfsdk:"debian_trivial_layout"`
	ChecksumPolicyType              types.String `tfsdk:"checksum_policy_type"`
	HandleReleases                  types.Bool   `tfsdk:"handle_releases"`
	HandleSnapshots                 types.Bool   `tfsdk:"handle_snapshots"`
	MaxUniqueSnapshots              types.Int64  `tfsdk:"max_unique_snapshots"`
	MaxUniqueTags                   types.Int64  `tfsdk:"max_unique_tags"`
	SnapshotVersionBehavior         types.String `tfsdk:"snapshot_version_behavior"`
	SupporessPomConsistencyChecks   types.Bool   `tfsdk:"supporess_pom_consistency_checks"`
	BlackOut                        types.Bool   `tfsdk:"blacked_out"`
	XrayIndex                       types.Bool   `tfsdk:"xray_index"`
	PropertySets                    types.Set    `tfsdk:"property_sets"`
	ArchiveBrowsingEnabled          types.Bool   `tfsdk:"archive_browsing_enabled"`
	CalculateYumMetadata            types.Bool   `tfsdk:"calculate_yum_metadata"`
	YumRootDepth                    types.Int64  `tfsdk:"yum_root_depth"`
	DockerApiVersion                types.String `tfsdk:"docker_api_version"`
	EnableFileListsIndexing         types.Bool   `tfsdk:"enable_file_lists_indexing"`
	OptionalIndexCompressionFormats types.Set    `tfsdk:"optional_index_compression_formats"`
	DownloadRedirect                types.Bool   `tfsdk:"download_direct"`
	CDNRedirect                     types.Bool   `tfsdk:"cdn_redirect"`
	BlockPushingSchema1             types.Bool   `tfsdk:"block_pushing_schema_1"`
	PrimaryKeyPairRef               types.String `tfsdk:"primary_key_pair_ref"`
	SecondaryKeyPairRef             types.String `tfsdk:"secondary_key_pair_ref"`
	PriorityResolution              types.Bool   `tfsdk:"priority_resolution"`
}

type AllLocalDataSourceModel struct {
	PackageType types.String `tfsdk:"package_type"`
	Repos       types.Set    `tfsdk:"repos"`
}

var repoAttrType = map[string]attr.Type{
	"debian_trivial_layout":              types.BoolType,
	"checksum_policy_type":               types.StringType,
	"handle_releases":                    types.BoolType,
	"handle_snapshots":                   types.BoolType,
	"max_unique_snapshots":               types.Int64Type,
	"max_unique_tags":                    types.Int64Type,
	"snapshot_version_behavior":          types.StringType,
	"supporess_pom_consistency_checks":   types.BoolType,
	"blacked_out":                        types.BoolType,
	"xray_index":                         types.BoolType,
	"property_sets":                      types.SetType{ElemType: types.StringType},
	"archive_browsing_enabled":           types.BoolType,
	"calculate_yum_metadata":             types.BoolType,
	"yum_root_depth":                     types.Int64Type,
	"docker_api_version":                 types.StringType,
	"enable_file_lists_indexing":         types.BoolType,
	"optional_index_compression_formats": types.SetType{ElemType: types.StringType},
	"download_direct":                    types.BoolType,
	"cdn_redirect":                       types.BoolType,
	"block_pushing_schema_1":             types.BoolType,
	"primary_key_pair_ref":               types.StringType,
	"secondary_key_pair_ref":             types.StringType,
	"priority_resolution":                types.BoolType,
}

func (m *AllLocalDataSourceModel) FromAPIModel(ctx context.Context, data []LocalAPIModel) diag.Diagnostics {

	var repos []attr.Value

	for _, d := range data {
		propertySets, diag := types.SetValueFrom(ctx, types.StringType, d.PropertySets)
		if diag != nil {
			return diag
		}

		optionalIndexCompressionFormats, diag := types.SetValueFrom(ctx, types.StringType, d.OptionalIndexCompressionFormats)
		if diag != nil {
			return diag
		}

		// dataSourceModel := repository.DataSourceModel{}
		// diag = dataSourceModel.FromAPIModel(ctx, d.APIModel)
		// if diag != nil {
		// 	return diag
		// }

		repo := types.ObjectValueMust(
			repoAttrType,
			map[string]attr.Value{
				"debian_trivial_layout":              types.BoolValue(d.DebianTrivialLayout),
				"checksum_policy_type":               types.StringValue(d.ChecksumPolicyType),
				"handle_releases":                    types.BoolValue(d.HandleReleases),
				"handle_snapshots":                   types.BoolValue(d.HandleSnapshots),
				"max_unique_snapshots":               types.Int64Value(d.MaxUniqueSnapshots),
				"max_unique_tags":                    types.Int64Value(d.MaxUniqueTags),
				"snapshot_version_behavior":          types.StringValue(d.SnapshotVersionBehavior),
				"supporess_pom_consistency_checks":   types.BoolValue(d.SupporessPomConsistencyChecks),
				"blacked_out":                        types.BoolValue(d.BlackOut),
				"xray_index":                         types.BoolValue(d.XrayIndex),
				"property_sets":                      propertySets,
				"archive_browsing_enabled":           types.BoolValue(d.ArchiveBrowsingEnabled),
				"calculate_yum_metadata":             types.BoolValue(d.CalculateYumMetadata),
				"yum_root_depth":                     types.Int64Value(d.YumRootDepth),
				"docker_api_version":                 types.StringValue(d.DockerApiVersion),
				"enable_file_lists_indexing":         types.BoolValue(d.EnableFileListsIndexing),
				"optional_index_compression_formats": optionalIndexCompressionFormats,
				"download_direct":                    types.BoolValue(d.DownloadRedirect),
				"cdn_redirect":                       types.BoolValue(d.CDNRedirect),
				"block_pushing_schema_1":             types.BoolValue(d.BlockPushingSchema1),
				"primary_key_pair_ref":               types.StringValue(d.PrimaryKeyPairRef),
				"secondary_key_pair_ref":             types.StringValue(d.SecondaryKeyPairRef),
				"priority_resolution":                types.BoolValue(d.PriorityResolution),
			},
		)

		repos = append(repos, repo)
	}

	reposSet, d := types.SetValue(types.ObjectType{AttrTypes: repoAttrType}, repos)
	if d != nil {
		return d
	}

	m.Repos = reposSet

	return nil
}

var allLocalReposSchema map[string]schema.Attribute = utilsdk.MergeMaps(
	repository.RepoSchema,
	map[string]schema.Attribute{
		"debian_trivial_layout":              schema.BoolAttribute{Computed: true},
		"checksum_policy_type":               schema.StringAttribute{Computed: true},
		"handle_releases":                    schema.BoolAttribute{Computed: true},
		"handle_snapshots":                   schema.BoolAttribute{Computed: true},
		"max_unique_snapshots":               schema.Int64Attribute{Computed: true},
		"max_unique_tags":                    schema.Int64Attribute{Computed: true},
		"snapshot_version_behavior":          schema.StringAttribute{Computed: true},
		"supporess_pom_consistency_checks":   schema.BoolAttribute{Computed: true},
		"black_out":                          schema.BoolAttribute{Computed: true},
		"xray_index":                         schema.BoolAttribute{Computed: true},
		"property_sets":                      schema.SetAttribute{ElementType: types.StringType, Computed: true},
		"archive_browsing_enabled":           schema.BoolAttribute{Computed: true},
		"calculate_yum_metadata":             schema.BoolAttribute{Computed: true},
		"yum_root_depth":                     schema.Int64Attribute{Computed: true},
		"docker_api_version":                 schema.StringAttribute{Computed: true},
		"enable_file_lists_indexing":         schema.BoolAttribute{Computed: true},
		"optional_index_compression_formats": schema.SetAttribute{ElementType: types.StringType, Computed: true},
		"download_direct":                    schema.BoolAttribute{Computed: true},
		"cdn_redirect":                       schema.BoolAttribute{Computed: true},
		"block_pushing_schema_1":             schema.BoolAttribute{Computed: true},
		"primary_key_pair_ref":               schema.StringAttribute{Computed: true},
		"secondary_key_pair_ref":             schema.StringAttribute{Computed: true},
		"priority_resolution":                schema.BoolAttribute{Computed: true},
	},
)

type LocalAPIModel struct {
	repository.APIModel
	DebianTrivialLayout             bool     `json:"debian_trivial_layout"`
	ChecksumPolicyType              string   `json:"checksum_policy_type"`
	HandleReleases                  bool     `json:"handle_releases"`
	HandleSnapshots                 bool     `json:"handle_snapshots"`
	MaxUniqueSnapshots              int64    `json:"max_unique_snapshots"`
	MaxUniqueTags                   int64    `json:"max_unique_tags"`
	SnapshotVersionBehavior         string   `json:"snapshot_version_behavior"`
	SupporessPomConsistencyChecks   bool     `json:"supporess_pom_consistency_checks"`
	BlackOut                        bool     `json:"blacked_out"`
	XrayIndex                       bool     `json:"xray_index"`
	PropertySets                    []string `json:"property_sets"`
	ArchiveBrowsingEnabled          bool     `json:"archive_browsing_enabled"`
	CalculateYumMetadata            bool     `json:"calculate_yum_metadata"`
	YumRootDepth                    int64    `json:"yum_root_depth"`
	DockerApiVersion                string   `json:"docker_api_version"`
	EnableFileListsIndexing         bool     `json:"enable_file_lists_indexing"`
	OptionalIndexCompressionFormats []string `json:"optional_index_compression_formats"`
	DownloadRedirect                bool     `json:"download_direct"`
	CDNRedirect                     bool     `json:"cdn_redirect"`
	BlockPushingSchema1             bool     `json:"block_pushing_schema_1"`
	PrimaryKeyPairRef               string   `json:"primary_key_pair_ref"`
	SecondaryKeyPairRef             string   `json:"secondary_key_pair_ref"`
	PriorityResolution              bool     `json:"priority_resolution"`
}

func (d *AllLocalDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "artifactory_local_all_repository"
}

func (d *AllLocalDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"package_type": schema.StringAttribute{
				Required: true,
			},
			"repos": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: allLocalReposSchema,
				},
			},
		},
	}
}

func (d *AllLocalDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	d.ProviderData = req.ProviderData.(utilsdk.ProvderMetadata)
}

func (d *AllLocalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AllLocalDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repos []LocalAPIModel
	_, err := d.ProviderData.Client.R().
		SetQueryParams(map[string]string{
			"repositoryType": "local",
			"packageType":    data.PackageType.ValueString(),
		}).
		SetResult(&repos).
		Get(repository.EndPoint)

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Data Source",
			"An unexpected error occurred while fetch the data source. "+
				"Please report this issue to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(data.FromAPIModel(ctx, repos)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
