package remote

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	"github.com/samber/lo"
)

// BaseRemoteRepositoryDataSourceModel contains common fields for all remote repository data sources
type BaseRemoteRepositoryDataSourceModel struct {
	repository.BaseRepositoryDataSourceModel
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
type BaseRemoteRepositoryAPIModel struct {
	repository.BaseRepositoryAPIModel
	URL                               string `json:"url"`
	Username                          string `json:"username"`
	Password                          string `json:"password"`
	Proxy                             string `json:"proxy"`
	DisableProxy                      bool   `json:"disableProxy"`
	RemoteRepoLayoutRef               string `json:"remoteRepoLayoutRef"`
	HardFail                          bool   `json:"hardFail"`
	Offline                           bool   `json:"offline"`
	QueryParams                       string `json:"queryParams"`
	StoreArtifactsLocally             bool   `json:"storeArtifactsLocally"`
	SocketTimeoutMillis               int64  `json:"socketTimeoutMillis"`
	LocalAddress                      string `json:"localAddress"`
	RetrievalCachePeriodSecs          int64  `json:"retrievalCachePeriodSecs"`
	MissedRetrievalCachePeriodSecs    int64  `json:"missedRetrievalCachePeriodSecs"`
	MetadataRetrievalTimeoutSecs      int64  `json:"metadataRetrievalTimeoutSecs"`
	UnusedArtifactsCleanupPeriodHours int64  `json:"unusedArtifactsCleanupPeriodHours"`
	AssumedOfflinePeriodSecs          int64  `json:"assumedOfflinePeriodSecs"`
	ShareConfiguration                bool   `json:"shareConfiguration"`
	SynchronizeProperties             bool   `json:"synchronizeProperties"`
	BlockMismatchingMimeTypes         bool   `json:"blockMismatchingMimeTypes"`
	AllowAnyHostAuth                  bool   `json:"allowAnyHostAuth"`
	EnableCookieManagement            bool   `json:"enableCookieManagement"`
	BypassHeadRequests                bool   `json:"bypassHeadRequests"`
	ClientTLSCertificate              string `json:"clientTlsCertificate"`
	MismatchingMimeTypeOverrideList   string `json:"mismatchingMimeTypesOverrideList"`
	ListRemoteFolderItems             bool   `json:"listRemoteFolderItems"`
	DisableURLNormalization           bool   `json:"disableUrlNormalization"`
}

// BaseRemoteSchemaAttributes returns the base schema attributes for all remote repository data sources
func BaseRemoteSchemaAttributes() map[string]schema.Attribute {
	return lo.Assign(
		repository.BaseSchemaAttributes(),
		map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The remote repo URL.",
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
				MarkdownDescription: "When set, the repository should store cached artifacts locally. When not set, artifacts are not stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server setups over a high-speed permanent connection, and is one of the flags to be tuned for a specific server in Artifactory that can have a significant impact on performance.",
				Computed:            true,
			},
			"socket_timeout_millis": schema.Int64Attribute{
				MarkdownDescription: "Network timeout (in ms) to use when establishing a connection and for unanswered requests. Timing out on a network operation is considered a retrieval failure.",
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
				MarkdownDescription: "When set, the repository should store cached artifacts locally. When not set, artifacts are not stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server setups over a high-speed permanent connection, and is one of the flags to be tuned for a specific server in Artifactory that can have a significant impact on performance.",
				Computed:            true,
			},
			"synchronize_properties": schema.BoolAttribute{
				MarkdownDescription: "When set, remote artifacts are downloaded along with their properties and metadata to the local repository as well.",
				Computed:            true,
			},
			"block_mismatching_mime_types": schema.BoolAttribute{
				MarkdownDescription: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even when downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.",
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
				MarkdownDescription: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even when downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.",
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

// CommonRemoteFromAPIModel provides common conversion logic from API model to Terraform model for remote repositories
func CommonRemoteFromAPIModel(ctx context.Context, baseModel *BaseRemoteRepositoryDataSourceModel, apiModel BaseRemoteRepositoryAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	// Convert common fields using the base repository function
	diags.Append(repository.CommonFromAPIModel(ctx, &baseModel.BaseRepositoryDataSourceModel, apiModel.BaseRepositoryAPIModel)...)
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
	baseModel.HardFail = types.BoolValue(apiModel.HardFail)
	baseModel.Offline = types.BoolValue(apiModel.Offline)
	baseModel.QueryParams = types.StringValue(apiModel.QueryParams)
	baseModel.StoreArtifactsLocally = types.BoolValue(apiModel.StoreArtifactsLocally)
	baseModel.SocketTimeoutMillis = types.Int64Value(apiModel.SocketTimeoutMillis)
	baseModel.LocalAddress = types.StringValue(apiModel.LocalAddress)
	baseModel.RetrievalCachePeriodSecs = types.Int64Value(apiModel.RetrievalCachePeriodSecs)
	baseModel.MissedRetrievalCachePeriodSecs = types.Int64Value(apiModel.MissedRetrievalCachePeriodSecs)
	baseModel.MetadataRetrievalTimeoutSecs = types.Int64Value(apiModel.MetadataRetrievalTimeoutSecs)
	baseModel.UnusedArtifactsCleanupPeriodHours = types.Int64Value(apiModel.UnusedArtifactsCleanupPeriodHours)
	baseModel.AssumedOfflinePeriodSecs = types.Int64Value(apiModel.AssumedOfflinePeriodSecs)
	baseModel.ShareConfiguration = types.BoolValue(apiModel.ShareConfiguration)
	baseModel.SynchronizeProperties = types.BoolValue(apiModel.SynchronizeProperties)
	baseModel.BlockMismatchingMimeTypes = types.BoolValue(apiModel.BlockMismatchingMimeTypes)
	baseModel.AllowAnyHostAuth = types.BoolValue(apiModel.AllowAnyHostAuth)
	baseModel.EnableCookieManagement = types.BoolValue(apiModel.EnableCookieManagement)
	baseModel.BypassHeadRequests = types.BoolValue(apiModel.BypassHeadRequests)
	baseModel.ClientTLSCertificate = types.StringValue(apiModel.ClientTLSCertificate)
	baseModel.MismatchingMimeTypeOverrideList = types.StringValue(apiModel.MismatchingMimeTypeOverrideList)
	baseModel.ListRemoteFolderItems = types.BoolValue(apiModel.ListRemoteFolderItems)
	baseModel.DisableURLNormalization = types.BoolValue(apiModel.DisableURLNormalization)

	return diags
}
