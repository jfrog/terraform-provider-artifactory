package remote

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	datasource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/samber/lo"
)

// Framework types and functions

// BaseRemoteRepositoryDataSourceModel contains common fields for all remote repository data sources
type BaseRemoteRepositoryDataSourceModel struct {
	datasource_repository.BaseRepositoryDataSourceModel
	URL                               types.String `tfsdk:"url"`
	Username                          types.String `tfsdk:"username"`
	Password                          types.String `tfsdk:"password"`
	Proxy                             types.String `tfsdk:"proxy"`
	DisableProxy                      types.Bool   `tfsdk:"disable_proxy"`
	RemoteRepoLayoutRef               types.String `tfsdk:"remote_repo_layout_ref"`
	HardFail                          types.Bool   `tfsdk:"hard_fail"`
	Offline                           types.Bool   `tfsdk:"offline"`
	QueryParams                       types.String `tfsdk:"query_params"`
	StoreArtifactsLocally             types.Bool   `tfsdk:"store_artifacts_locally"`
	SocketTimeoutMillis               types.Int64  `tfsdk:"socket_timeout_millis"`
	LocalAddress                      types.String `tfsdk:"local_address"`
	RetrievalCachePeriodSecs          types.Int64  `tfsdk:"retrieval_cache_period_seconds"`
	MissedRetrievalCachePeriodSecs    types.Int64  `tfsdk:"missed_cache_period_seconds"`
	MetadataRetrievalTimeoutSecs      types.Int64  `tfsdk:"metadata_retrieval_timeout_secs"`
	UnusedArtifactsCleanupPeriodHours types.Int64  `tfsdk:"unused_artifacts_cleanup_period_hours"`
	AssumedOfflinePeriodSecs          types.Int64  `tfsdk:"assumed_offline_period_secs"`
	ShareConfiguration                types.Bool   `tfsdk:"share_configuration"`
	SynchronizeProperties             types.Bool   `tfsdk:"synchronize_properties"`
	BlockMismatchingMimeTypes         types.Bool   `tfsdk:"block_mismatching_mime_types"`
	AllowAnyHostAuth                  types.Bool   `tfsdk:"allow_any_host_auth"`
	EnableCookieManagement            types.Bool   `tfsdk:"enable_cookie_management"`
	BypassHeadRequests                types.Bool   `tfsdk:"bypass_head_requests"`
	ClientTLSCertificate              types.String `tfsdk:"client_tls_certificate"`
	MismatchingMimeTypeOverrideList   types.String `tfsdk:"mismatching_mime_types_override_list"`
	ListRemoteFolderItems             types.Bool   `tfsdk:"list_remote_folder_items"`
	DisableURLNormalization           types.Bool   `tfsdk:"disable_url_normalization"`
}

// BaseRemoteRepositoryAPIModel contains common fields for all remote repository API models
// This is a helper type for conversion - actual API models embed remote.RemoteAPIModel
type BaseRemoteRepositoryAPIModel struct {
	datasource_repository.BaseRepositoryAPIModel
	remote.RemoteAPIModel
}

// BaseRemoteSchemaAttributes returns the base schema attributes for all remote repository data sources
func BaseRemoteSchemaAttributes() map[string]schema.Attribute {
	return lo.Assign(
		datasource_repository.BaseDataSourceAttributes,
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "This is a URL to the remote registry. Consider using HTTPS to ensure a secure connection.",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username to use when connecting to the remote repository.",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password to use when connecting to the remote repository.",
				Computed:            true,
				Sensitive:           true,
			},
			"proxy": schema.StringAttribute{
				MarkdownDescription: "Proxy key from Artifactory Proxies settings.",
				Computed:            true,
			},
			"disable_proxy": schema.BoolAttribute{
				MarkdownDescription: "When set, this repository will not use the Artifactory proxy configuration.",
				Computed:            true,
			},
			"remote_repo_layout_ref": schema.StringAttribute{
				MarkdownDescription: "Repository layout key for the remote layout mapping.",
				Computed:            true,
			},
			"hard_fail": schema.BoolAttribute{
				MarkdownDescription: "When set, Artifactory will return an error to the client that causes the build to fail if there is a failure to communicate with this repository.",
				Computed:            true,
			},
			"offline": schema.BoolAttribute{
				MarkdownDescription: "When set, the repository is treated as a Remote repository. The default value is `true`.",
				Computed:            true,
			},
			"query_params": schema.StringAttribute{
				MarkdownDescription: "Query parameters to include in the request to the remote repository.",
				Computed:            true,
			},
			"store_artifacts_locally": schema.BoolAttribute{
				MarkdownDescription: "When set, the repository should store cached artifacts locally.",
				Computed:            true,
			},
			"socket_timeout_millis": schema.Int64Attribute{
				MarkdownDescription: "Network timeout (in ms) to use when establishing a connection and for unanswered requests.",
				Computed:            true,
			},
			"local_address": schema.StringAttribute{
				MarkdownDescription: "The network address of the local machine that should be used to connect to the remote repository system.",
				Computed:            true,
			},
			"retrieval_cache_period_seconds": schema.Int64Attribute{
				MarkdownDescription: "The metadataRetrievalCachePeriod (in seconds) specifies how long the cache metadata should be considered valid.",
				Computed:            true,
			},
			"missed_cache_period_seconds": schema.Int64Attribute{
				MarkdownDescription: "The number of seconds to cache artifact retrieval misses (artifact not found).",
				Computed:            true,
			},
			"metadata_retrieval_timeout_secs": schema.Int64Attribute{
				MarkdownDescription: "The number of seconds to wait for an artifact to be downloaded before considering the download failed.",
				Computed:            true,
			},
			"unused_artifacts_cleanup_period_hours": schema.Int64Attribute{
				MarkdownDescription: "The number of hours to cache items that were not accessed.",
				Computed:            true,
			},
			"assumed_offline_period_secs": schema.Int64Attribute{
				MarkdownDescription: "The number of seconds the repository should wait before checking whether a remote VCS repository has been changed.",
				Computed:            true,
			},
			"share_configuration": schema.BoolAttribute{
				MarkdownDescription: "When set, the repository should store cached artifacts locally.",
				Computed:            true,
			},
			"synchronize_properties": schema.BoolAttribute{
				MarkdownDescription: "When set, remote artifacts are downloaded along with their properties and metadata to the local repository as well.",
				Computed:            true,
			},
			"block_mismatching_mime_types": schema.BoolAttribute{
				MarkdownDescription: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource.",
				Computed:            true,
			},
			"allow_any_host_auth": schema.BoolAttribute{
				MarkdownDescription: "Also known as 'Lenient Host Authentication', Allow credentials of this repository to be used on requests redirected to any other host.",
				Computed:            true,
			},
			"enable_cookie_management": schema.BoolAttribute{
				MarkdownDescription: "Enables cookie management if the remote repository uses cookies to manage client state.",
				Computed:            true,
			},
			"bypass_head_requests": schema.BoolAttribute{
				MarkdownDescription: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource.",
				Computed:            true,
			},
			"client_tls_certificate": schema.StringAttribute{
				MarkdownDescription: "Client TLS certificate name.",
				Computed:            true,
			},
			"mismatching_mime_types_override_list": schema.StringAttribute{
				MarkdownDescription: "Lists of artifacts that should be accepted by the proxy repository.",
				Computed:            true,
			},
			"list_remote_folder_items": schema.BoolAttribute{
				MarkdownDescription: "When set, Artifactory will list remote folder items when listing the repository contents.",
				Computed:            true,
			},
			"disable_url_normalization": schema.BoolAttribute{
				MarkdownDescription: "When set, Artifactory will not normalize URLs when downloading artifacts.",
				Computed:            true,
			},
		},
	)
}

// RemoteDataSourceAttributes defines the attributes for remote repository datasources
var RemoteDataSourceAttributes = BaseRemoteSchemaAttributes

// CommonRemoteFromAPIModel provides common conversion logic from API model to Terraform model for remote repositories
func CommonRemoteFromAPIModel(ctx context.Context, baseModel *BaseRemoteRepositoryDataSourceModel, apiModel BaseRemoteRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Convert common fields using the base repository function
	diags.Append(datasource_repository.CommonFromAPIModel(ctx, &baseModel.BaseRepositoryDataSourceModel, apiModel.BaseRepositoryAPIModel)...)
	if diags.HasError() {
		return diags
	}

	// Convert remote-specific fields
	baseModel.URL = types.StringValue(apiModel.URL)
	baseModel.Username = types.StringValue(apiModel.Username)
	baseModel.Password = types.StringValue(apiModel.Password)
	baseModel.Proxy = types.StringValue(apiModel.Proxy)
	baseModel.DisableProxy = types.BoolValue(apiModel.DisableProxy)
	baseModel.RemoteRepoLayoutRef = types.StringValue(apiModel.RemoteRepoLayoutRef)
	if apiModel.HardFail != nil {
		baseModel.HardFail = types.BoolValue(*apiModel.HardFail)
	} else {
		baseModel.HardFail = types.BoolNull()
	}
	if apiModel.Offline != nil {
		baseModel.Offline = types.BoolValue(*apiModel.Offline)
	} else {
		baseModel.Offline = types.BoolNull()
	}
	baseModel.QueryParams = types.StringValue(apiModel.QueryParams)
	if apiModel.StoreArtifactsLocally != nil {
		baseModel.StoreArtifactsLocally = types.BoolValue(*apiModel.StoreArtifactsLocally)
	} else {
		baseModel.StoreArtifactsLocally = types.BoolNull()
	}
	baseModel.SocketTimeoutMillis = types.Int64Value(apiModel.SocketTimeoutMillis)
	baseModel.LocalAddress = types.StringValue(apiModel.LocalAddress)
	baseModel.RetrievalCachePeriodSecs = types.Int64Value(apiModel.RetrievalCachePeriodSecs)
	baseModel.MissedRetrievalCachePeriodSecs = types.Int64Value(apiModel.MissedRetrievalCachePeriodSecs)
	baseModel.MetadataRetrievalTimeoutSecs = types.Int64Value(apiModel.MetadataRetrievalTimeoutSecs)
	baseModel.UnusedArtifactsCleanupPeriodHours = types.Int64Value(apiModel.UnusedArtifactsCleanupPeriodHours)
	baseModel.AssumedOfflinePeriodSecs = types.Int64Value(apiModel.AssumedOfflinePeriodSecs)
	if apiModel.ShareConfiguration != nil {
		baseModel.ShareConfiguration = types.BoolValue(*apiModel.ShareConfiguration)
	} else {
		baseModel.ShareConfiguration = types.BoolNull()
	}
	if apiModel.SynchronizeProperties != nil {
		baseModel.SynchronizeProperties = types.BoolValue(*apiModel.SynchronizeProperties)
	} else {
		baseModel.SynchronizeProperties = types.BoolNull()
	}
	if apiModel.BlockMismatchingMimeTypes != nil {
		baseModel.BlockMismatchingMimeTypes = types.BoolValue(*apiModel.BlockMismatchingMimeTypes)
	} else {
		baseModel.BlockMismatchingMimeTypes = types.BoolNull()
	}
	if apiModel.AllowAnyHostAuth != nil {
		baseModel.AllowAnyHostAuth = types.BoolValue(*apiModel.AllowAnyHostAuth)
	} else {
		baseModel.AllowAnyHostAuth = types.BoolNull()
	}
	if apiModel.EnableCookieManagement != nil {
		baseModel.EnableCookieManagement = types.BoolValue(*apiModel.EnableCookieManagement)
	} else {
		baseModel.EnableCookieManagement = types.BoolNull()
	}
	if apiModel.BypassHeadRequests != nil {
		baseModel.BypassHeadRequests = types.BoolValue(*apiModel.BypassHeadRequests)
	} else {
		baseModel.BypassHeadRequests = types.BoolNull()
	}
	baseModel.ClientTLSCertificate = types.StringValue(apiModel.ClientTLSCertificate)
	baseModel.MismatchingMimeTypeOverrideList = types.StringValue(apiModel.MismatchingMimeTypeOverrideList)
	baseModel.ListRemoteFolderItems = types.BoolValue(apiModel.ListRemoteFolderItems)
	baseModel.DisableURLNormalization = types.BoolValue(apiModel.DisableURLNormalization)

	return diags
}

// SDKv2 types and functions

var getSchema = func(schemas map[int16]map[string]*sdkv2_schema.Schema) map[string]*sdkv2_schema.Schema {
	s := schemas[remote.CurrentSchemaVersion]

	s["url"].Required = false
	s["url"].Optional = true

	return s
}

var VcsRemoteRepoSchemaSDKv2 = map[string]*sdkv2_schema.Schema{
	"vcs_git_provider": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		Default:          "GITHUB",
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GITHUB", "BITBUCKET", "OLDSTASH", "STASH", "ARTIFACTORY", "CUSTOM"}, false)),
		Description:      `Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "GITHUB".`,
	},
	"vcs_git_download_url": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      `This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.`,
	},
}
