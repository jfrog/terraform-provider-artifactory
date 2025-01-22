package remote

import (
	"context"
	"fmt"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkv2_validator "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	utilvalidator "github.com/jfrog/terraform-provider-shared/validator"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const (
	Rclass               = "remote"
	CurrentSchemaVersion = 3
)

func NewRemoteRepositoryResource(packageType, packageName string, resourceModelType, apiModelType reflect.Type) remoteResource {
	return remoteResource{
		BaseResource: repository.NewRepositoryResource(packageType, packageName, Rclass, resourceModelType, apiModelType),
	}
}

type remoteResource struct {
	repository.BaseResource
}

type RemoteResourceModel struct {
	local.LocalResourceModel
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
	ContentSynchronisation            types.List   `tfsdk:"content_synchronisation"`
	MismatchingMimeTypeOverrideList   types.String `tfsdk:"mismatching_mime_types_override_list"`
	ListRemoteFolderItems             types.Bool   `tfsdk:"list_remote_folder_items"`
	CDNRedirect                       types.Bool   `tfsdk:"cdn_redirect"`
	DisableURLNormalization           types.Bool   `tfsdk:"disable_url_normalization"`
}

type vcsResourceModel struct {
	VCSGitProvider    types.String `tfsdk:"vcs_git_provider"`
	VCSGitDownloadURL types.String `tfsdk:"vcs_git_download_url"`
}

func (r *RemoteResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r RemoteResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *RemoteResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *RemoteResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r RemoteResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r RemoteResourceModel) ToAPIModel(ctx context.Context, packageType string) (RemoteAPIModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}
	localRepositoryAPIModel := model.(local.LocalAPIModel)
	localRepositoryAPIModel.Rclass = Rclass

	if r.RepoLayoutRef.IsNull() {
		repoLayoutRef, err := repository.GetDefaultRepoLayoutRef(Rclass, packageType)
		if err != nil {
			diags.AddError(
				"Failed to get default repo layout ref",
				err.Error(),
			)
		}
		localRepositoryAPIModel.RepoLayoutRef = repoLayoutRef
	} else {
		localRepositoryAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()
	}

	var contentSynchronisation *ContentSynchronisation
	elems := r.ContentSynchronisation.Elements()
	if len(elems) > 0 {
		attrs := elems[0].(types.Object).Attributes()
		contentSynchronisation = &ContentSynchronisation{
			Enabled: attrs["enabled"].(types.Bool).ValueBool(),
			Statistics: ContentSynchronisationStatistics{
				Enabled: attrs["statistics_enabled"].(types.Bool).ValueBool(),
			},
			Properties: ContentSynchronisationProperties{
				Enabled: attrs["properties_enabled"].(types.Bool).ValueBool(),
			},
			Source: ContentSynchronisationSource{
				OriginAbsenceDetection: attrs["source_origin_absence_detection"].(types.Bool).ValueBool(),
			},
		}
	}

	return RemoteAPIModel{
		LocalAPIModel:                     localRepositoryAPIModel,
		URL:                               r.URL.ValueString(),
		Username:                          r.Username.ValueString(),
		Password:                          r.Password.ValueString(),
		DisableProxy:                      r.DisableProxy.ValueBool(),
		RemoteRepoLayoutRef:               r.RemoteRepoLayoutRef.ValueString(),
		HardFail:                          r.HardFail.ValueBoolPointer(),
		Offline:                           r.Offline.ValueBoolPointer(),
		QueryParams:                       r.QueryParams.ValueString(),
		StoreArtifactsLocally:             r.StoreArtifactsLocally.ValueBoolPointer(),
		SocketTimeoutMillis:               r.SocketTimeoutMillis.ValueInt64(),
		LocalAddress:                      r.LocalAddress.ValueString(),
		RetrievalCachePeriodSecs:          r.RetrievalCachePeriodSecs.ValueInt64(),
		MissedRetrievalCachePeriodSecs:    r.MissedRetrievalCachePeriodSecs.ValueInt64(),
		MetadataRetrievalTimeoutSecs:      r.MetadataRetrievalTimeoutSecs.ValueInt64(),
		UnusedArtifactsCleanupPeriodHours: r.UnusedArtifactsCleanupPeriodHours.ValueInt64(),
		AssumedOfflinePeriodSecs:          r.AssumedOfflinePeriodSecs.ValueInt64(),
		SynchronizeProperties:             r.SynchronizeProperties.ValueBoolPointer(),
		BlockMismatchingMimeTypes:         r.BlockMismatchingMimeTypes.ValueBoolPointer(),
		AllowAnyHostAuth:                  r.AllowAnyHostAuth.ValueBoolPointer(),
		EnableCookieManagement:            r.EnableCookieManagement.ValueBoolPointer(),
		BypassHeadRequests:                r.BypassHeadRequests.ValueBoolPointer(),
		ClientTLSCertificate:              r.ClientTLSCertificate.ValueStringPointer(),
		ContentSynchronisation:            contentSynchronisation,
		MismatchingMimeTypeOverrideList:   r.MismatchingMimeTypeOverrideList.ValueString(),
		ListRemoteFolderItems:             r.ListRemoteFolderItems.ValueBool(),
		CDNRedirect:                       r.CDNRedirect.ValueBool(),
		DisableURLNormalization:           r.DisableURLNormalization.ValueBool(),
	}, diags
}

var contentSynchronisationAttrType = types.ObjectType{
	AttrTypes: contentSynchronisationAttrTypes,
}

var contentSynchronisationAttrTypes = map[string]attr.Type{
	"enabled":                         types.BoolType,
	"statistics_enabled":              types.BoolType,
	"properties_enabled":              types.BoolType,
	"source_origin_absence_detection": types.BoolType,
}

func (r *RemoteResourceModel) FromAPIModel(ctx context.Context, apiModel RemoteAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.LocalResourceModel.FromAPIModel(ctx, apiModel.LocalAPIModel)

	r.URL = types.StringValue(apiModel.URL)
	r.Username = types.StringValue(apiModel.Username)
	r.Proxy = types.StringValue(apiModel.Proxy)
	r.DisableProxy = types.BoolValue(apiModel.DisableProxy)
	r.RemoteRepoLayoutRef = types.StringValue(apiModel.RemoteRepoLayoutRef)
	r.HardFail = types.BoolPointerValue(apiModel.HardFail)
	r.Offline = types.BoolPointerValue(apiModel.Offline)
	r.QueryParams = types.StringValue(apiModel.QueryParams)
	r.StoreArtifactsLocally = types.BoolPointerValue(apiModel.StoreArtifactsLocally)
	r.SocketTimeoutMillis = types.Int64Value(apiModel.SocketTimeoutMillis)
	r.LocalAddress = types.StringValue(apiModel.LocalAddress)
	r.RetrievalCachePeriodSecs = types.Int64Value(apiModel.RetrievalCachePeriodSecs)
	r.MissedRetrievalCachePeriodSecs = types.Int64Value(apiModel.MissedRetrievalCachePeriodSecs)
	r.MetadataRetrievalTimeoutSecs = types.Int64Value(apiModel.MetadataRetrievalTimeoutSecs)
	r.UnusedArtifactsCleanupPeriodHours = types.Int64Value(apiModel.UnusedArtifactsCleanupPeriodHours)
	r.ShareConfiguration = types.BoolPointerValue(apiModel.ShareConfiguration)
	r.AssumedOfflinePeriodSecs = types.Int64Value(apiModel.AssumedOfflinePeriodSecs)
	r.SynchronizeProperties = types.BoolPointerValue(apiModel.SynchronizeProperties)
	r.BlockMismatchingMimeTypes = types.BoolPointerValue(apiModel.BlockMismatchingMimeTypes)
	r.AllowAnyHostAuth = types.BoolPointerValue(apiModel.AllowAnyHostAuth)
	r.EnableCookieManagement = types.BoolPointerValue(apiModel.EnableCookieManagement)
	r.BypassHeadRequests = types.BoolPointerValue(apiModel.BypassHeadRequests)
	r.ClientTLSCertificate = types.StringPointerValue(apiModel.ClientTLSCertificate)

	contentSynchronisationList := types.ListNull(contentSynchronisationAttrType)
	// only update plan/state with ContentSynchronisation from API if it is enabled
	// not perfect conditional but we are limited by SDKv2 loosy logic for 'computed' list
	// if apiModel.ContentSynchronisation != nil && apiModel.ContentSynchronisation.Enabled {
	if apiModel.ContentSynchronisation != nil && apiModel.ContentSynchronisation.Enabled {
		contentSynchronisation, ds := types.ObjectValue(
			contentSynchronisationAttrTypes,
			map[string]attr.Value{
				"enabled":                         types.BoolValue(apiModel.ContentSynchronisation.Enabled),
				"statistics_enabled":              types.BoolValue(apiModel.ContentSynchronisation.Statistics.Enabled),
				"properties_enabled":              types.BoolValue(apiModel.ContentSynchronisation.Properties.Enabled),
				"source_origin_absence_detection": types.BoolValue(apiModel.ContentSynchronisation.Source.OriginAbsenceDetection),
			},
		)
		if ds.HasError() {
			diags.Append(ds...)
		}

		contentSynchronisationList, ds = types.ListValue(
			contentSynchronisationAttrType,
			[]attr.Value{contentSynchronisation},
		)

		if ds != nil {
			diags = append(diags, ds...)
		}
	}
	r.ContentSynchronisation = contentSynchronisationList

	r.MismatchingMimeTypeOverrideList = types.StringValue(apiModel.MismatchingMimeTypeOverrideList)
	r.ListRemoteFolderItems = types.BoolValue(apiModel.ListRemoteFolderItems)
	r.CDNRedirect = types.BoolValue(apiModel.CDNRedirect)
	r.DisableURLNormalization = types.BoolValue(apiModel.DisableURLNormalization)

	return diags
}

type RemoteAPIModel struct {
	local.LocalAPIModel
	URL                               string                  `json:"url"`
	Username                          string                  `json:"username"`
	Password                          string                  `json:"password,omitempty"` // must have 'omitempty' to avoid sending an empty string on update, if attribute is ignored by the provider.
	Proxy                             string                  `json:"proxy"`
	DisableProxy                      bool                    `json:"disableProxy"`
	RemoteRepoLayoutRef               string                  `json:"remoteRepoLayoutRef"`
	HardFail                          *bool                   `json:"hardFail,omitempty"`
	Offline                           *bool                   `json:"offline,omitempty"`
	QueryParams                       string                  `json:"queryParams,omitempty"`
	StoreArtifactsLocally             *bool                   `json:"storeArtifactsLocally,omitempty"`
	SocketTimeoutMillis               int64                   `json:"socketTimeoutMillis"`
	LocalAddress                      string                  `json:"localAddress"`
	RetrievalCachePeriodSecs          int64                   `json:"retrievalCachePeriodSecs"`
	MissedRetrievalCachePeriodSecs    int64                   `json:"missedRetrievalCachePeriodSecs"`
	MetadataRetrievalTimeoutSecs      int64                   `json:"metadataRetrievalTimeoutSecs"`
	UnusedArtifactsCleanupPeriodHours int64                   `json:"unusedArtifactsCleanupPeriodHours"`
	AssumedOfflinePeriodSecs          int64                   `json:"assumedOfflinePeriodSecs"`
	ShareConfiguration                *bool                   `json:"shareConfiguration,omitempty"`
	SynchronizeProperties             *bool                   `json:"synchronizeProperties"`
	BlockMismatchingMimeTypes         *bool                   `json:"blockMismatchingMimeTypes"`
	AllowAnyHostAuth                  *bool                   `json:"allowAnyHostAuth,omitempty"`
	EnableCookieManagement            *bool                   `json:"enableCookieManagement,omitempty"`
	BypassHeadRequests                *bool                   `json:"bypassHeadRequests,omitempty"`
	ClientTLSCertificate              *string                 `json:"clientTlsCertificate,omitempty"`
	ContentSynchronisation            *ContentSynchronisation `json:"contentSynchronisation,omitempty"`
	MismatchingMimeTypeOverrideList   string                  `json:"mismatchingMimeTypesOverrideList"`
	ListRemoteFolderItems             bool                    `json:"listRemoteFolderItems"`
	CDNRedirect                       bool                    `json:"cdnRedirect"`
	DisableURLNormalization           bool                    `json:"disableUrlNormalization"`
}

type vcsAPIModel struct {
	GitProvider    *string `json:"vcsGitProvider,omitempty"`
	GitDownloadURL *string `json:"vcsGitDownloadUrl,omitempty"`
}

var RemoteAttributes = lo.Assign(
	local.LocalAttributes,
	map[string]schema.Attribute{
		"url": schema.StringAttribute{
			Required: true,
			Validators: []validator.String{
				validatorfw_string.IsURLHttpOrHttps(),
			},
			MarkdownDescription: "This is a URL to the remote registry. Consider using HTTPS to ensure a secure connection.",
		},
		"username": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(""),
		},
		"password": schema.StringAttribute{
			Optional:  true,
			Sensitive: true,
		},
		"proxy": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			MarkdownDescription: "Proxy key from Artifactory Proxies settings. Can't be set if `disable_proxy = true`.",
		},
		"disable_proxy": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set to `true`, the proxy is disabled, and not returned in the API response body. If there is a default proxy set for the Artifactory instance, it will be ignored, too. Introduced since Artifactory 7.41.7.",
		},
		"remote_repo_layout_ref": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			MarkdownDescription: "Repository layout key for the remote layout mapping. Repository can be created without this attribute (or set to an empty string). Once it's set, it can't be removed by passing an empty string or removing the attribute, that will be ignored by the Artifactory API. UI shows an error message, if the user tries to remove the value.",
		},
		"hard_fail": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			MarkdownDescription: "When set, Artifactory will return an error to the client that causes the build to fail if there " +
				"is a failure to communicate with this repository.",
		},
		"offline": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "If set, Artifactory does not try to fetch remote artifacts. Only locally-cached artifacts are retrieved.",
		},
		"blacked_out": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "(A.K.A 'Ignore Repository' on the UI) When set, the repository or its local cache do not participate in artifact resolution.",
		},
		"store_artifacts_locally": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(true),
			MarkdownDescription: "When set, the repository should store cached artifacts locally. When not set, artifacts are not " +
				"stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server " +
				"setups over a high-speed LAN, with one Artifactory caching certain data on central storage, and streaming " +
				"it directly to satellite pass-though Artifactory servers.",
		},
		"socket_timeout_millis": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(15000),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			MarkdownDescription: "Network timeout (in ms) to use when establishing a connection and for unanswered requests. " +
				"Timing out on a network operation is considered a retrieval failure.",
		},
		"local_address": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(""),
			MarkdownDescription: "The local address to be used when creating connections. " +
				"Useful for specifying the interface to use on systems with multiple network interfaces.",
		},
		"retrieval_cache_period_seconds": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(7200),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			MarkdownDescription: "Metadata Retrieval Cache Period (Sec) in the UI. This value refers to the number of seconds to cache " +
				"metadata files before checking for newer versions on remote server. A value of 0 indicates no caching.",
		},
		"metadata_retrieval_timeout_secs": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(60),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			MarkdownDescription: "Metadata Retrieval Cache Timeout (Sec) in the UI.This value refers to the number of seconds to wait " +
				"for retrieval from the remote before serving locally cached artifact or fail the request.",
		},
		"missed_cache_period_seconds": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(1800),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			MarkdownDescription: "Missed Retrieval Cache Period (Sec) in the UI. The number of seconds to cache artifact retrieval " +
				"misses (artifact not found). A value of 0 indicates no caching.",
		},
		"unused_artifacts_cleanup_period_hours": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(0),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			MarkdownDescription: "Unused Artifacts Cleanup Period (Hr) in the UI. The number of hours to wait before an artifact is " +
				"deemed 'unused' and eligible for cleanup from the repository. A value of 0 means automatic cleanup of cached artifacts is disabled.",
		},
		"assumed_offline_period_secs": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(300),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
			MarkdownDescription: "The number of seconds the repository stays in assumed offline state after a connection error. " +
				"At the end of this time, an online check is attempted in order to reset the offline status. " +
				"A value of 0 means the repository is never assumed offline.",
		},
		"share_configuration": schema.BoolAttribute{
			Optional:           true,
			Computed:           true,
			Default:            booldefault.StaticBool(false),
			DeprecationMessage: "No longer supported",
		},
		"synchronize_properties": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, remote artifacts are fetched along with their properties.",
		},
		"block_mismatching_mime_types": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(true),
			MarkdownDescription: "If set, artifacts will fail to download if a mismatch is detected between requested and received " +
				"mimetype, according to the list specified in the system properties file under blockedMismatchingMimeTypes. " +
				"You can override by adding mimetypes to the override list 'mismatching_mime_types_override_list'.",
		},
		"allow_any_host_auth": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "'Lenient Host Authentication' in the UI. Allow credentials of this repository to be used on requests redirected to any other host.",
		},
		"enable_cookie_management": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enables cookie management if the remote repository uses cookies to manage client state.",
		},
		"bypass_head_requests": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			MarkdownDescription: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. " +
				"In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the " +
				"artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.",
		},
		"client_tls_certificate": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: "Client TLS certificate name.",
		},
		"query_params": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(""),
			MarkdownDescription: "Custom HTTP query parameters that will be automatically included in all remote resource requests. " +
				"For example: `param1=val1&param2=val2&param3=val3`",
		},
		"list_remote_folder_items": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			MarkdownDescription: "Lists the items of remote folders in simple and list browsing. The remote content is cached " +
				"according to the value of the 'Retrieval Cache Period'. Default value is 'false'. This field exists in the API but not in the UI.",
		},
		"mismatching_mime_types_override_list": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(""),
			Validators: []validator.String{
				stringvalidator.RegexMatches(regexp.MustCompile(`.+(?:,.+)*`), "must be comma separated string"),
			},
			MarkdownDescription: "The set of mime types that should override the block_mismatching_mime_types setting. " +
				"Eg: 'application/json,application/xml'. Default value is empty.",
		},
		"download_direct": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact " +
				"directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only. Default value is 'false'.",
		},
		"cdn_redirect": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is 'false'",
		},
		"disable_url_normalization": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Whether to disable URL normalization. Default is `false`.",
		},
	},
)

var remoteBlocks = map[string]schema.Block{
	"content_synchronisation": schema.ListNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"enabled": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: "If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.",
				},
				"statistics_enabled": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: "If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.",
				},
				"properties_enabled": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: "If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.",
				},
				"source_origin_absence_detection": schema.BoolAttribute{
					Optional:            true,
					Computed:            true,
					Default:             booldefault.StaticBool(false),
					MarkdownDescription: "If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'",
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeBetween(0, 1),
		},
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
	},
}

var vcsAttributes = map[string]schema.Attribute{
	"vcs_git_provider": schema.StringAttribute{
		Optional: true,
		Computed: true,
		Default:  stringdefault.StaticString("GITHUB"),
		Validators: []validator.String{
			stringvalidator.OneOf("GITHUB", "BITBUCKET", "OLDSTASH", "STASH", "ARTIFACTORY", "CUSTOM"),
		},
		MarkdownDescription: `Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "GITHUB".`,
	},
	"vcs_git_download_url": schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
		MarkdownDescription: `This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.`,
	},
}

// SDKv2
type ContentSynchronisation struct {
	Enabled    bool                             `json:"enabled"`
	Statistics ContentSynchronisationStatistics `json:"statistics"`
	Properties ContentSynchronisationProperties `json:"properties"`
	Source     ContentSynchronisationSource     `json:"source"`
}

type ContentSynchronisationStatistics struct {
	Enabled bool `hcl:"statistics_enabled" json:"enabled"`
}

type ContentSynchronisationProperties struct {
	Enabled bool `hcl:"properties_enabled" json:"enabled"`
}

type ContentSynchronisationSource struct {
	OriginAbsenceDetection bool `hcl:"source_origin_absence_detection" json:"originAbsenceDetection"`
}

type RepositoryRemoteBaseParams struct {
	Key                               string                  `json:"key,omitempty"`
	ProjectKey                        string                  `json:"projectKey"`
	ProjectEnvironments               []string                `json:"environments"`
	Rclass                            string                  `json:"rclass"`
	PackageType                       string                  `json:"packageType,omitempty"`
	Url                               string                  `json:"url"`
	Username                          string                  `json:"username"`
	Password                          string                  `json:"password,omitempty"` // must have 'omitempty' to avoid sending an empty string on update, if attribute is ignored by the provider.
	Proxy                             string                  `json:"proxy"`
	DisableProxy                      bool                    `json:"disableProxy"`
	Description                       string                  `json:"description"`
	Notes                             string                  `json:"notes"`
	IncludesPattern                   string                  `json:"includesPattern"`
	ExcludesPattern                   string                  `json:"excludesPattern"`
	RepoLayoutRef                     string                  `json:"repoLayoutRef"`
	RemoteRepoLayoutRef               string                  `json:"remoteRepoLayoutRef"`
	HardFail                          *bool                   `json:"hardFail,omitempty"`
	Offline                           *bool                   `json:"offline,omitempty"`
	BlackedOut                        *bool                   `json:"blackedOut,omitempty"`
	XrayIndex                         bool                    `json:"xrayIndex"`
	QueryParams                       string                  `json:"queryParams,omitempty"`
	PriorityResolution                bool                    `json:"priorityResolution"`
	StoreArtifactsLocally             *bool                   `json:"storeArtifactsLocally,omitempty"`
	SocketTimeoutMillis               int                     `json:"socketTimeoutMillis"`
	LocalAddress                      string                  `json:"localAddress"`
	RetrievalCachePeriodSecs          int                     `hcl:"retrieval_cache_period_seconds" json:"retrievalCachePeriodSecs"`
	MissedRetrievalCachePeriodSecs    int                     `hcl:"missed_cache_period_seconds" json:"missedRetrievalCachePeriodSecs"`
	MetadataRetrievalTimeoutSecs      int                     `json:"metadataRetrievalTimeoutSecs"`
	UnusedArtifactsCleanupPeriodHours int                     `json:"unusedArtifactsCleanupPeriodHours"`
	AssumedOfflinePeriodSecs          int                     `hcl:"assumed_offline_period_secs" json:"assumedOfflinePeriodSecs"`
	ShareConfiguration                *bool                   `hcl:"share_configuration" json:"shareConfiguration,omitempty"`
	SynchronizeProperties             *bool                   `hcl:"synchronize_properties" json:"synchronizeProperties"`
	BlockMismatchingMimeTypes         *bool                   `hcl:"block_mismatching_mime_types" json:"blockMismatchingMimeTypes"`
	PropertySets                      []string                `hcl:"property_sets" json:"propertySets,omitempty"`
	AllowAnyHostAuth                  *bool                   `hcl:"allow_any_host_auth" json:"allowAnyHostAuth,omitempty"`
	EnableCookieManagement            *bool                   `hcl:"enable_cookie_management" json:"enableCookieManagement,omitempty"`
	BypassHeadRequests                *bool                   `hcl:"bypass_head_requests" json:"bypassHeadRequests,omitempty"`
	ClientTLSCertificate              string                  `hcl:"client_tls_certificate" json:"clientTlsCertificate,omitempty"`
	ContentSynchronisation            *ContentSynchronisation `hcl:"content_synchronisation" json:"contentSynchronisation,omitempty"`
	MismatchingMimeTypeOverrideList   string                  `hcl:"mismatching_mime_types_override_list" json:"mismatchingMimeTypesOverrideList"`
	ListRemoteFolderItems             bool                    `json:"listRemoteFolderItems"`
	DownloadRedirect                  bool                    `hcl:"download_direct" json:"downloadRedirect,omitempty"`
	CdnRedirect                       bool                    `json:"cdnRedirect"`
	DisableURLNormalization           bool                    `hcl:"disable_url_normalization" json:"disableUrlNormalization"`
	ArchiveBrowsingEnabled            *bool                   `json:"archiveBrowsingEnabled,omitempty"`
}

func (r RepositoryRemoteBaseParams) GetRclass() string {
	return r.Rclass
}

type JavaRemoteRepo struct {
	RepositoryRemoteBaseParams
	FetchJarsEagerly             bool   `json:"fetchJarsEagerly"`
	FetchSourcesEagerly          bool   `json:"fetchSourcesEagerly"`
	RemoteRepoChecksumPolicyType string `json:"remoteRepoChecksumPolicyType"`
	HandleReleases               bool   `json:"handleReleases"`
	HandleSnapshots              bool   `json:"handleSnapshots"`
	SuppressPomConsistencyChecks bool   `json:"suppressPomConsistencyChecks"`
	RejectInvalidJars            bool   `json:"rejectInvalidJars"`
	MaxUniqueSnapshots           int    `json:"maxUniqueSnapshots"`
}

type RepositoryVcsParams struct {
	VcsGitProvider    string `json:"vcsGitProvider"`
	VcsGitDownloadUrl string `json:"vcsGitDownloadUrl"`
}

func (bp RepositoryRemoteBaseParams) Id() string {
	return bp.Key
}

var PackageTypesLikeBasic = []string{
	repository.AlpinePackageType,
	repository.ChefPackageType,
	repository.CondaPackageType,
	repository.CranPackageType,
	repository.DebianPackageType,
	repository.GitLFSPackageType,
	repository.OpkgPackageType,
	repository.P2PackageType,
	repository.PubPackageType,
	repository.PuppetPackageType,
	repository.RPMPackageType,
	repository.SwiftPackageType,
}

var BaseSchema = lo.Assign(
	repository.ProxySchemaSDKv2,
	map[string]*sdkv2_schema.Schema{
		"url": {
			Type:         sdkv2_schema.TypeString,
			Required:     true,
			ValidateFunc: sdkv2_validator.IsURLWithHTTPorHTTPS,
			Description:  "This is a URL to the remote registry. Consider using HTTPS to ensure a secure connection.",
		},
		"username": {
			Type:     sdkv2_schema.TypeString,
			Optional: true,
		},
		"password": {
			Type:      sdkv2_schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"description": {
			Type:     sdkv2_schema.TypeString,
			Optional: true,
			DiffSuppressFunc: func(_, old, new string, _ *sdkv2_schema.ResourceData) bool {
				// this is literally what comes back from the server
				return old == fmt.Sprintf("%s (local file cache)", new)
			},
			Description: "Public description.",
		},
		"remote_repo_layout_ref": {
			Type:        sdkv2_schema.TypeString,
			Optional:    true,
			Description: "Repository layout key for the remote layout mapping. Repository can be created without this attribute (or set to an empty string). Once it's set, it can't be removed by passing an empty string or removing the attribute, that will be ignored by the Artifactory API. UI shows an error message, if the user tries to remove the value.",
		},
		"hard_fail": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "When set, Artifactory will return an error to the client that causes the build to fail if there " +
				"is a failure to communicate with this repository.",
		},
		"offline": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If set, Artifactory does not try to fetch remote artifacts. Only locally-cached artifacts are retrieved.",
		},
		"blacked_out": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "(A.K.A 'Ignore Repository' on the UI) When set, the repository or its local cache do not participate in artifact resolution.",
		},
		"xray_index": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "Enable Indexing In Xray. Repository will be indexed with the default retention period. " +
				"You will be able to change it via Xray settings.",
		},
		"store_artifacts_locally": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  true,
			Description: "When set, the repository should store cached artifacts locally. When not set, artifacts are not " +
				"stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server " +
				"setups over a high-speed LAN, with one Artifactory caching certain data on central storage, and streaming " +
				"it directly to satellite pass-though Artifactory servers.",
		},
		"socket_timeout_millis": {
			Type:         sdkv2_schema.TypeInt,
			Optional:     true,
			Default:      15000,
			ValidateFunc: sdkv2_validator.IntAtLeast(0),
			Description: "Network timeout (in ms) to use when establishing a connection and for unanswered requests. " +
				"Timing out on a network operation is considered a retrieval failure.",
		},
		"local_address": {
			Type:     sdkv2_schema.TypeString,
			Optional: true,
			Description: "The local address to be used when creating connections. " +
				"Useful for specifying the interface to use on systems with multiple network interfaces.",
		},
		"retrieval_cache_period_seconds": {
			Type:         sdkv2_schema.TypeInt,
			Optional:     true,
			Default:      7200,
			ValidateFunc: sdkv2_validator.IntAtLeast(0),
			Description: "Metadata Retrieval Cache Period (Sec) in the UI. This value refers to the number of seconds to cache " +
				"metadata files before checking for newer versions on remote server. A value of 0 indicates no caching.",
		},
		"metadata_retrieval_timeout_secs": {
			Type:         sdkv2_schema.TypeInt,
			Optional:     true,
			Default:      60,
			ValidateFunc: sdkv2_validator.IntAtLeast(0),
			Description: "Metadata Retrieval Cache Timeout (Sec) in the UI.This value refers to the number of seconds to wait " +
				"for retrieval from the remote before serving locally cached artifact or fail the request.",
		},
		"missed_cache_period_seconds": {
			Type:         sdkv2_schema.TypeInt,
			Optional:     true,
			Default:      1800,
			ValidateFunc: sdkv2_validator.IntAtLeast(0),
			Description: "Missed Retrieval Cache Period (Sec) in the UI. The number of seconds to cache artifact retrieval " +
				"misses (artifact not found). A value of 0 indicates no caching.",
		},
		"unused_artifacts_cleanup_period_hours": {
			Type:         sdkv2_schema.TypeInt,
			Optional:     true,
			Default:      0,
			ValidateFunc: sdkv2_validator.IntAtLeast(0),
			Description: "Unused Artifacts Cleanup Period (Hr) in the UI. The number of hours to wait before an artifact is " +
				"deemed 'unused' and eligible for cleanup from the repository. A value of 0 means automatic cleanup of cached artifacts is disabled.",
		},
		"assumed_offline_period_secs": {
			Type:         sdkv2_schema.TypeInt,
			Optional:     true,
			Default:      300,
			ValidateFunc: sdkv2_validator.IntAtLeast(0),
			Description: "The number of seconds the repository stays in assumed offline state after a connection error. " +
				"At the end of this time, an online check is attempted in order to reset the offline status. " +
				"A value of 0 means the repository is never assumed offline.",
		},
		// There is no corresponding field in the UI, but the attribute is returned by Get, default is 'false'.
		"share_configuration": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Computed: true,
		},
		"synchronize_properties": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, remote artifacts are fetched along with their properties.",
		},
		// Default value in UI is 'true', at the same time if the repo was created with API, the default is 'false'.
		// We are repeating the UI behavior.
		"block_mismatching_mime_types": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  true,
			Description: "If set, artifacts will fail to download if a mismatch is detected between requested and received " +
				"mimetype, according to the list specified in the system properties file under blockedMismatchingMimeTypes. " +
				"You can override by adding mimetypes to the override list 'mismatching_mime_types_override_list'.",
		},
		"property_sets": {
			Type:        sdkv2_schema.TypeSet,
			Elem:        &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
			Set:         sdkv2_schema.HashString,
			Optional:    true,
			Description: "List of property set names",
		},
		"allow_any_host_auth": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "'Lenient Host Authentication' in the UI. Allow credentials of this repository to be used on requests redirected to any other host.",
		},
		"enable_cookie_management": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Enables cookie management if the remote repository uses cookies to manage client state.",
		},
		"bypass_head_requests": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. " +
				"In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the " +
				"artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.",
		},
		"priority_resolution": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "Setting Priority Resolution takes precedence over the resolution order when resolving virtual " +
				"repositories. Setting repositories with priority will cause metadata to be merged only from repositories " +
				"set with a priority. If a package is not found in those repositories, Artifactory will merge from repositories marked as non-priority.",
		},
		"client_tls_certificate": {
			Type:        sdkv2_schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Client TLS certificate name.",
		},
		"content_synchronisation": {
			Type:     sdkv2_schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &sdkv2_schema.Resource{
				Schema: map[string]*sdkv2_schema.Schema{
					"enabled": {
						Type:        sdkv2_schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.",
					},
					"statistics_enabled": {
						Type:        sdkv2_schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.",
					},
					"properties_enabled": {
						Type:        sdkv2_schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.",
					},
					"source_origin_absence_detection": {
						Type:        sdkv2_schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'",
					},
				},
			},
		},
		"query_params": {
			Type:     sdkv2_schema.TypeString,
			Optional: true,
			Description: "Custom HTTP query parameters that will be automatically included in all remote resource requests. " +
				"For example: `param1=val1&param2=val2&param3=val3`",
		},
		"list_remote_folder_items": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "Lists the items of remote folders in simple and list browsing. The remote content is cached " +
				"according to the value of the 'Retrieval Cache Period'. Default value is 'false'. This field exists in the API but not in the UI.",
		},
		"mismatching_mime_types_override_list": {
			Type:             sdkv2_schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: utilvalidator.CommaSeperatedList,
			StateFunc:        utilsdk.FormatCommaSeparatedString,
			Description: "The set of mime types that should override the block_mismatching_mime_types setting. " +
				"Eg: 'application/json,application/xml'. Default value is empty.",
		},
		"download_direct": {
			Type:     sdkv2_schema.TypeBool,
			Optional: true,
			Default:  false,
			Description: "When set, download requests to this repository will redirect the client to download the artifact " +
				"directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only. Default value is 'false'.",
		},
		"cdn_redirect": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is 'false'",
		},
		"disable_url_normalization": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to disable URL normalization, default is `false`.",
		},
		"archive_browsing_enabled": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, you may view content such as HTML or Javadoc files directly from Artifactory.\nThis may not be safe and therefore requires strict content moderation to prevent malicious users from uploading content that may compromise security (e.g., cross-site scripting attacks).",
		},
	},
)

var baseSchemaV1 = lo.Assign(
	repository.BaseSchemaV1,
	BaseSchema,
	map[string]*sdkv2_schema.Schema{
		"propagate_query_params": {
			Type:        sdkv2_schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
		},
	},
)

var baseSchemaV2 = lo.Assign(
	repository.BaseSchemaV1,
	BaseSchema,
)

var baseSchemaV3 = lo.Assign(
	repository.BaseSchemaV1,
	BaseSchema,
)

var GetSchemas = func(s map[string]*sdkv2_schema.Schema) map[int16]map[string]*sdkv2_schema.Schema {
	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(
			baseSchemaV1,
			s,
		),
		1: lo.Assign(
			baseSchemaV1,
			s,
		),
		2: lo.Assign(
			baseSchemaV2,
			s,
		),
		3: lo.Assign(
			baseSchemaV3,
			s,
		),
	}
}

var VcsRemoteRepoSchemaSDKv2 = map[string]*sdkv2_schema.Schema{
	"vcs_git_provider": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		Default:          "GITHUB",
		ValidateDiagFunc: sdkv2_validator.ToDiagFunc(sdkv2_validator.StringInSlice([]string{"GITHUB", "BITBUCKET", "OLDSTASH", "STASH", "ARTIFACTORY", "CUSTOM"}, false)),
		Description:      `Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "GITHUB".`,
	},
	"vcs_git_download_url": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: sdkv2_validator.ToDiagFunc(sdkv2_validator.StringIsNotEmpty),
		Description:      `This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.`,
	},
}

func JavaSchema(packageType string, suppressPom bool) map[string]*sdkv2_schema.Schema {
	return lo.Assign(
		BaseSchema,
		map[string]*sdkv2_schema.Schema{
			"fetch_jars_eagerly": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `When set, if a POM is requested, Artifactory attempts to fetch the corresponding jar in the background. This will accelerate first access time to the jar when it is subsequently requested. Default value is 'false'.`,
			},
			"fetch_sources_eagerly": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `When set, if a binaries jar is requested, Artifactory attempts to fetch the corresponding source jar in the background. This will accelerate first access time to the source jar when it is subsequently requested. Default value is 'false'.`,
			},
			"remote_repo_checksum_policy_type": {
				Type:     sdkv2_schema.TypeString,
				Optional: true,
				Default:  "generate-if-absent",
				ValidateDiagFunc: sdkv2_validator.ToDiagFunc(sdkv2_validator.StringInSlice([]string{
					"generate-if-absent",
					"fail",
					"ignore-and-generate",
					"pass-thru",
				}, false)),
				Description: `Checking the Checksum effectively verifies the integrity of a deployed resource. The Checksum Policy determines how the system behaves when a client checksum for a remote resource is missing or conflicts with the locally calculated checksum. Default value is 'generate-if-absent'.`,
			},
			"handle_releases": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `If set, Artifactory allows you to deploy release artifacts into this repository. Default value is 'true'.`,
			},
			"handle_snapshots": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `If set, Artifactory allows you to deploy snapshot artifacts into this repository. Default value is 'true'.`,
			},
			"suppress_pom_consistency_checks": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     suppressPom,
				Description: `By default, the system keeps your repositories healthy by refusing POMs with incorrect coordinates (path). If the groupId:artifactId:version information inside the POM does not match the deployed path, Artifactory rejects the deployment with a "409 Conflict" error. You can disable this behavior by setting this attribute to 'true'. Default value is 'false'.`,
			},
			"reject_invalid_jars": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `Reject the caching of jar files that are found to be invalid. For example, pseudo jars retrieved behind a "captive portal". Default value is 'false'.`,
			},
			"max_unique_snapshots": {
				Type:             sdkv2_schema.TypeInt,
				Optional:         true,
				Default:          0,
				ValidateDiagFunc: sdkv2_validator.ToDiagFunc(sdkv2_validator.IntAtLeast(0)),
				Description: "The maximum number of unique snapshots of a single artifact to store. Once the number of " +
					"snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is " +
					"no limit, and unique snapshots are not cleaned up.",
			},
		},
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)
}

func UnpackBaseRemoteRepo(s *sdkv2_schema.ResourceData, packageType string) RepositoryRemoteBaseParams {
	d := &utilsdk.ResourceData{ResourceData: s}

	repo := RepositoryRemoteBaseParams{
		Rclass:                            Rclass,
		Key:                               d.GetString("key", false),
		ProjectKey:                        d.GetString("project_key", false),
		ProjectEnvironments:               d.GetSet("project_environments"),
		PackageType:                       packageType, // must be set independently
		Url:                               d.GetString("url", false),
		Username:                          d.GetString("username", false),
		Password:                          d.GetString("password", false),
		Proxy:                             d.GetString("proxy", false),
		DisableProxy:                      d.GetBool("disable_proxy", false),
		Description:                       d.GetString("description", false),
		Notes:                             d.GetString("notes", false),
		IncludesPattern:                   d.GetString("includes_pattern", false),
		ExcludesPattern:                   d.GetString("excludes_pattern", false),
		RepoLayoutRef:                     d.GetString("repo_layout_ref", false),
		RemoteRepoLayoutRef:               d.GetString("remote_repo_layout_ref", false),
		HardFail:                          d.GetBoolRef("hard_fail", false),
		Offline:                           d.GetBoolRef("offline", false),
		BlackedOut:                        d.GetBoolRef("blacked_out", false),
		XrayIndex:                         d.GetBool("xray_index", false),
		DownloadRedirect:                  d.GetBool("download_direct", false),
		CdnRedirect:                       d.GetBool("cdn_redirect", false),
		QueryParams:                       d.GetString("query_params", false),
		StoreArtifactsLocally:             d.GetBoolRef("store_artifacts_locally", false),
		SocketTimeoutMillis:               d.GetInt("socket_timeout_millis", false),
		LocalAddress:                      d.GetString("local_address", false),
		RetrievalCachePeriodSecs:          d.GetInt("retrieval_cache_period_seconds", false),
		MissedRetrievalCachePeriodSecs:    d.GetInt("missed_cache_period_seconds", false),
		MetadataRetrievalTimeoutSecs:      d.GetInt("metadata_retrieval_timeout_secs", false),
		UnusedArtifactsCleanupPeriodHours: d.GetInt("unused_artifacts_cleanup_period_hours", false),
		AssumedOfflinePeriodSecs:          d.GetInt("assumed_offline_period_secs", false),
		ShareConfiguration:                d.GetBoolRef("share_configuration", false),
		SynchronizeProperties:             d.GetBoolRef("synchronize_properties", false),
		BlockMismatchingMimeTypes:         d.GetBoolRef("block_mismatching_mime_types", false),
		PropertySets:                      d.GetSet("property_sets"),
		AllowAnyHostAuth:                  d.GetBoolRef("allow_any_host_auth", false),
		EnableCookieManagement:            d.GetBoolRef("enable_cookie_management", false),
		BypassHeadRequests:                d.GetBoolRef("bypass_head_requests", false),
		ClientTLSCertificate:              d.GetString("client_tls_certificate", false),
		PriorityResolution:                d.GetBool("priority_resolution", false),
		ListRemoteFolderItems:             d.GetBool("list_remote_folder_items", false),
		MismatchingMimeTypeOverrideList:   d.GetString("mismatching_mime_types_override_list", false),
		DisableURLNormalization:           d.GetBool("disable_url_normalization", false),
		ArchiveBrowsingEnabled:            d.GetBoolRef("archive_browsing_enabled", false),
	}

	if v, ok := d.GetOk("content_synchronisation"); ok {
		contentSynchronisationConfig := v.([]interface{})[0].(map[string]interface{})
		enabled := contentSynchronisationConfig["enabled"].(bool)
		statisticsEnabled := contentSynchronisationConfig["statistics_enabled"].(bool)
		propertiesEnabled := contentSynchronisationConfig["properties_enabled"].(bool)
		sourceOriginAbsenceDetection := contentSynchronisationConfig["source_origin_absence_detection"].(bool)
		repo.ContentSynchronisation = &ContentSynchronisation{
			Enabled: enabled,
			Statistics: ContentSynchronisationStatistics{
				Enabled: statisticsEnabled,
			},
			Properties: ContentSynchronisationProperties{
				Enabled: propertiesEnabled,
			},
			Source: ContentSynchronisationSource{
				OriginAbsenceDetection: sourceOriginAbsenceDetection,
			},
		}
	}
	return repo
}

func UnpackVcsRemoteRepo(s *sdkv2_schema.ResourceData) RepositoryVcsParams {
	d := &utilsdk.ResourceData{ResourceData: s}
	return RepositoryVcsParams{
		VcsGitProvider:    d.GetString("vcs_git_provider", false),
		VcsGitDownloadUrl: d.GetString("vcs_git_download_url", false),
	}
}

func UnpackJavaRemoteRepo(s *sdkv2_schema.ResourceData, repoType string) JavaRemoteRepo {
	d := &utilsdk.ResourceData{ResourceData: s}
	return JavaRemoteRepo{
		RepositoryRemoteBaseParams:   UnpackBaseRemoteRepo(s, repoType),
		FetchJarsEagerly:             d.GetBool("fetch_jars_eagerly", false),
		FetchSourcesEagerly:          d.GetBool("fetch_sources_eagerly", false),
		RemoteRepoChecksumPolicyType: d.GetString("remote_repo_checksum_policy_type", false),
		HandleReleases:               d.GetBool("handle_releases", false),
		HandleSnapshots:              d.GetBool("handle_snapshots", false),
		SuppressPomConsistencyChecks: d.GetBool("suppress_pom_consistency_checks", false),
		RejectInvalidJars:            d.GetBool("reject_invalid_jars", false),
		MaxUniqueSnapshots:           d.GetInt("max_unique_snapshots", false),
	}
}

func mkResourceSchema(skeemas map[int16]map[string]*sdkv2_schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor repository.Constructor) *sdkv2_schema.Resource {
	var reader = repository.MkRepoRead(packer, constructor)

	return &sdkv2_schema.Resource{
		CreateContext: repository.MkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: repository.MkRepoUpdate(unpack, reader),
		DeleteContext: repository.DeleteRepo,

		Importer: &sdkv2_schema.ResourceImporter{
			StateContext: sdkv2_schema.ImportStatePassthroughContext,
		},

		StateUpgraders: []sdkv2_schema.StateUpgrader{
			{
				Type:    repository.Resource(skeemas[1]).CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceStateUpgradeV1,
				Version: 1,
			},
			{
				Type:    repository.Resource(skeemas[2]).CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceStateUpgradeV2,
				Version: 2,
			},
		},

		Schema:        skeemas[CurrentSchemaVersion],
		SchemaVersion: CurrentSchemaVersion,

		CustomizeDiff: customdiff.All(
			repository.ProjectEnvironmentsDiff,
			verifyExternalDependenciesDockerAndHelm,
			repository.VerifyDisableProxy,
			verifyRemoteRepoLayoutRef,
		),
	}
}

func verifyExternalDependenciesDockerAndHelm(_ context.Context, diff *sdkv2_schema.ResourceDiff, _ interface{}) error {
	// Skip the verification if sdkv2_schema doesn't have `external_dependencies_enabled` attribute (only docker and helm have it)
	if _, ok := diff.GetOkExists("external_dependencies_enabled"); !ok {
		return nil
	}
	for _, dep := range diff.Get("external_dependencies_patterns").([]interface{}) {
		if dep == "" {
			return fmt.Errorf("`external_dependencies_patterns` can't have an item of \"\" inside a list")
		}
	}

	if diff.Get("external_dependencies_enabled") == true {
		if _, ok := diff.GetOk("external_dependencies_patterns"); !ok {
			return fmt.Errorf("if `external_dependencies_enabled` is set to `true`, `external_dependencies_patterns` list must be set")
		}
	}

	return nil
}

func ResourceStateUpgradeV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState["package_type"] != "generic" {
		delete(rawState, "propagate_query_params")
	}

	return rawState, nil
}

func ResourceStateUpgradeV2(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	// this only works because the sdkv2_schema hasn't changed, except the removal of default value
	// from `project_key` attribute.
	if rawState["project_key"] == "default" {
		rawState["project_key"] = ""
	}

	if _, ok := rawState["archive_browsing_enabled"]; !ok {
		rawState["archive_browsing_enabled"] = false
	}

	if _, ok := rawState["disable_proxy"]; !ok {
		rawState["disable_proxy"] = false
	}

	if _, ok := rawState["disable_url_normalization"]; !ok {
		rawState["disable_url_normalization"] = false
	}

	if _, ok := rawState["property_sets"]; !ok {
		rawState["property_sets"] = []string{}
	}

	return rawState, nil
}

func verifyRemoteRepoLayoutRef(_ context.Context, diff *sdkv2_schema.ResourceDiff, _ interface{}) error {
	ref := diff.Get("remote_repo_layout_ref").(string)
	isChanged := diff.HasChange("remote_repo_layout_ref")

	if isChanged && len(ref) == 0 {
		return fmt.Errorf("empty remote_repo_layout_ref will not remove the actual attribute value and will be ignored by the API, " +
			"thus will create a state drift on the next plan. Please add the attribute, according to the repository type")
	}

	return nil
}
