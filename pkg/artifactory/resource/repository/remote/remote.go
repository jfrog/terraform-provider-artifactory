package remote

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

const rclass = "remote"

type RepositoryRemoteBaseParams struct {
	Key                               string                             `json:"key,omitempty"`
	ProjectKey                        string                             `json:"projectKey"`
	ProjectEnvironments               []string                           `json:"environments"`
	Rclass                            string                             `json:"rclass"`
	PackageType                       string                             `json:"packageType,omitempty"`
	Url                               string                             `json:"url"`
	Username                          string                             `json:"username"`
	Password                          string                             `json:"password,omitempty"` // must have 'omitempty' to avoid sending an empty string on update, if attribute is ignored by the provider.
	Proxy                             string                             `json:"proxy"`
	DisableProxy                      bool                               `json:"disableProxy"`
	Description                       string                             `json:"description"`
	Notes                             string                             `json:"notes"`
	IncludesPattern                   string                             `json:"includesPattern"`
	ExcludesPattern                   string                             `json:"excludesPattern"`
	RepoLayoutRef                     string                             `json:"repoLayoutRef"`
	RemoteRepoLayoutRef               string                             `json:"remoteRepoLayoutRef"`
	HardFail                          *bool                              `json:"hardFail,omitempty"`
	Offline                           *bool                              `json:"offline,omitempty"`
	BlackedOut                        *bool                              `json:"blackedOut,omitempty"`
	XrayIndex                         bool                               `json:"xrayIndex"`
	QueryParams                       string                             `json:"queryParams,omitempty"`
	PriorityResolution                bool                               `json:"priorityResolution"`
	StoreArtifactsLocally             *bool                              `json:"storeArtifactsLocally,omitempty"`
	SocketTimeoutMillis               int                                `json:"socketTimeoutMillis"`
	LocalAddress                      string                             `json:"localAddress"`
	RetrievalCachePeriodSecs          int                                `hcl:"retrieval_cache_period_seconds" json:"retrievalCachePeriodSecs"`
	MissedRetrievalCachePeriodSecs    int                                `hcl:"missed_cache_period_seconds" json:"missedRetrievalCachePeriodSecs"`
	MetadataRetrievalTimeoutSecs      int                                `json:"metadataRetrievalTimeoutSecs"`
	UnusedArtifactsCleanupPeriodHours int                                `json:"unusedArtifactsCleanupPeriodHours"`
	AssumedOfflinePeriodSecs          int                                `hcl:"assumed_offline_period_secs" json:"assumedOfflinePeriodSecs"`
	ShareConfiguration                *bool                              `hcl:"share_configuration" json:"shareConfiguration,omitempty"`
	SynchronizeProperties             *bool                              `hcl:"synchronize_properties" json:"synchronizeProperties"`
	BlockMismatchingMimeTypes         *bool                              `hcl:"block_mismatching_mime_types" json:"blockMismatchingMimeTypes"`
	PropertySets                      []string                           `hcl:"property_sets" json:"propertySets,omitempty"`
	AllowAnyHostAuth                  *bool                              `hcl:"allow_any_host_auth" json:"allowAnyHostAuth,omitempty"`
	EnableCookieManagement            *bool                              `hcl:"enable_cookie_management" json:"enableCookieManagement,omitempty"`
	BypassHeadRequests                *bool                              `hcl:"bypass_head_requests" json:"bypassHeadRequests,omitempty"`
	ClientTlsCertificate              string                             `hcl:"client_tls_certificate" json:"clientTlsCertificate,omitempty"`
	ContentSynchronisation            *repository.ContentSynchronisation `hcl:"content_synchronisation" json:"contentSynchronisation,omitempty"`
	MismatchingMimeTypeOverrideList   string                             `hcl:"mismatching_mime_types_override_list" json:"mismatchingMimeTypesOverrideList"`
	ListRemoteFolderItems             bool                               `json:"listRemoteFolderItems"`
	DownloadRedirect                  bool                               `hcl:"download_direct" json:"downloadRedirect,omitempty"`
	CdnRedirect                       bool                               `json:"cdnRedirect"`
	DisableURLNormalization           bool                               `hcl:"disable_url_normalization" json:"disableUrlNormalization"`
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
}

type RepositoryVcsParams struct {
	VcsGitProvider    string `json:"vcsGitProvider"`
	VcsGitDownloadUrl string `json:"vcsGitDownloadUrl"`
}

func (bp RepositoryRemoteBaseParams) Id() string {
	return bp.Key
}

var PackageTypesLikeBasic = []string{
	"alpine",
	"chef",
	"conda",
	"cran",
	"debian",
	"gems",
	"gitlfs",
	"opkg",
	"p2",
	"pub",
	"puppet",
	"rpm",
	"swift",
}

var BaseRemoteRepoSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		repository.BaseRepoSchema,
		repository.ProxySchema,
		map[string]*schema.Schema{
			"url": {
				Type:         schema.TypeString,
				Required:     isResource,
				Optional:     !isResource,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Description:  "The remote repo URL.",
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					// this is literally what comes back from the server
					return old == fmt.Sprintf("%s (local file cache)", new)
				},
				Description: "Public description.",
			},
			"remote_repo_layout_ref": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Repository layout key for the remote layout mapping. Repository can be created without this attribute (or set to an empty string). Once it's set, it can't be removed by passing an empty string or removing the attribute, that will be ignored by the Artifactory API. UI shows an error message, if the user tries to remove the value.",
			},
			"hard_fail": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "When set, Artifactory will return an error to the client that causes the build to fail if there " +
					"is a failure to communicate with this repository.",
			},
			"offline": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If set, Artifactory does not try to fetch remote artifacts. Only locally-cached artifacts are retrieved.",
			},
			"blacked_out": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "(A.K.A 'Ignore Repository' on the UI) When set, the repository or its local cache do not participate in artifact resolution.",
			},
			"xray_index": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Enable Indexing In Xray. Repository will be indexed with the default retention period. " +
					"You will be able to change it via Xray settings.",
			},
			"store_artifacts_locally": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "When set, the repository should store cached artifacts locally. When not set, artifacts are not " +
					"stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server " +
					"setups over a high-speed LAN, with one Artifactory caching certain data on central storage, and streaming " +
					"it directly to satellite pass-though Artifactory servers.",
			},
			"socket_timeout_millis": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      15000,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "Network timeout (in ms) to use when establishing a connection and for unanswered requests. " +
					"Timing out on a network operation is considered a retrieval failure.",
			},
			"local_address": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The local address to be used when creating connections. " +
					"Useful for specifying the interface to use on systems with multiple network interfaces.",
			},
			"retrieval_cache_period_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      7200,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "Metadata Retrieval Cache Period (Sec) in the UI. This value refers to the number of seconds to cache " +
					"metadata files before checking for newer versions on remote server. A value of 0 indicates no caching.",
			},
			"metadata_retrieval_timeout_secs": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "Metadata Retrieval Cache Timeout (Sec) in the UI.This value refers to the number of seconds to wait " +
					"for retrieval from the remote before serving locally cached artifact or fail the request.",
			},
			"missed_cache_period_seconds": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1800,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "Missed Retrieval Cache Period (Sec) in the UI. The number of seconds to cache artifact retrieval " +
					"misses (artifact not found). A value of 0 indicates no caching.",
			},
			"unused_artifacts_cleanup_period_hours": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "Unused Artifacts Cleanup Period (Hr) in the UI. The number of hours to wait before an artifact is " +
					"deemed 'unused' and eligible for cleanup from the repository. A value of 0 means automatic cleanup of cached artifacts is disabled.",
			},
			"assumed_offline_period_secs": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntAtLeast(0),
				Description: "The number of seconds the repository stays in assumed offline state after a connection error. " +
					"At the end of this time, an online check is attempted in order to reset the offline status. " +
					"A value of 0 means the repository is never assumed offline.",
			},
			// There is no corresponding field in the UI, but the attribute is returned by Get, default is 'false'.
			"share_configuration": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"synchronize_properties": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When set, remote artifacts are fetched along with their properties.",
			},
			// Default value in UI is 'true', at the same time if the repo was created with API, the default is 'false'.
			// We are repeating the UI behavior.
			"block_mismatching_mime_types": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "If set, artifacts will fail to download if a mismatch is detected between requested and received " +
					"mimetype, according to the list specified in the system properties file under blockedMismatchingMimeTypes. " +
					"You can override by adding mimetypes to the override list 'mismatching_mime_types_override_list'.",
			},
			"property_sets": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Description: "List of property set names",
			},
			"allow_any_host_auth": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "'Lenient Host Authentication' in the UI. Allow credentials of this repository to be used on requests redirected to any other host.",
			},
			"enable_cookie_management": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Enables cookie management if the remote repository uses cookies to manage client state.",
			},
			"bypass_head_requests": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. " +
					"In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the " +
					"artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.",
			},
			"priority_resolution": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Setting Priority Resolution takes precedence over the resolution order when resolving virtual " +
					"repositories. Setting repositories with priority will cause metadata to be merged only from repositories " +
					"set with a priority. If a package is not found in those repositories, Artifactory will merge from repositories marked as non-priority.",
			},
			"client_tls_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Client TLS certificate name.",
			},
			"content_synchronisation": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.",
						},
						"statistics_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.",
						},
						"properties_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.",
						},
						"source_origin_absence_detection": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'",
						},
					},
				},
			},
			"query_params": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Custom HTTP query parameters that will be automatically included in all remote resource requests. " +
					"For example: `param1=val1&param2=val2&param3=val3`",
			},
			"list_remote_folder_items": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				Description: "Lists the items of remote folders in simple and list browsing. The remote content is cached " +
					"according to the value of the 'Retrieval Cache Period'. Default value is 'true'.",
			},
			"mismatching_mime_types_override_list": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validator.CommaSeperatedList,
				StateFunc:        utilsdk.FormatCommaSeparatedString,
				Description: "The set of mime types that should override the block_mismatching_mime_types setting. " +
					"Eg: 'application/json,application/xml'. Default value is empty.",
			},
			"download_direct": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "When set, download requests to this repository will redirect the client to download the artifact " +
					"directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only. Default value is 'false'.",
			},
			"cdn_redirect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is 'false'",
			},
			"disable_url_normalization": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to disable URL normalization, default is `false`.",
			},
		},
	)
}

var baseRemoteRepoSchemaV1 = utilsdk.MergeMaps(
	BaseRemoteRepoSchema(true),
	map[string]*schema.Schema{
		"propagate_query_params": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
		},
	},
)

var baseRemoteRepoSchemaV2 = BaseRemoteRepoSchema(true)

var VcsRemoteRepoSchema = map[string]*schema.Schema{
	"vcs_git_provider": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          "GITHUB",
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GITHUB", "BITBUCKET", "OLDSTASH", "STASH", "ARTIFACTORY", "CUSTOM"}, false)),
		Description:      `Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "GITHUB".`,
	},
	"vcs_git_download_url": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      `This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.`,
	},
}

func JavaRemoteSchema(isResource bool, packageType string, suppressPom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"fetch_jars_eagerly": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `When set, if a POM is requested, Artifactory attempts to fetch the corresponding jar in the background. This will accelerate first access time to the jar when it is subsequently requested. Default value is 'false'.`,
			},
			"fetch_sources_eagerly": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `When set, if a binaries jar is requested, Artifactory attempts to fetch the corresponding source jar in the background. This will accelerate first access time to the source jar when it is subsequently requested. Default value is 'false'.`,
			},
			"remote_repo_checksum_policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "generate-if-absent",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					"generate-if-absent",
					"fail",
					"ignore-and-generate",
					"pass-thru",
				}, false)),
				Description: `Checking the Checksum effectively verifies the integrity of a deployed resource. The Checksum Policy determines how the system behaves when a client checksum for a remote resource is missing or conflicts with the locally calculated checksum. Default value is 'generate-if-absent'.`,
			},
			"handle_releases": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `If set, Artifactory allows you to deploy release artifacts into this repository. Default value is 'true'.`,
			},
			"handle_snapshots": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: `If set, Artifactory allows you to deploy snapshot artifacts into this repository. Default value is 'true'.`,
			},
			"suppress_pom_consistency_checks": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     suppressPom,
				Description: `By default, the system keeps your repositories healthy by refusing POMs with incorrect coordinates (path). If the groupId:artifactId:version information inside the POM does not match the deployed path, Artifactory rejects the deployment with a "409 Conflict" error. You can disable this behavior by setting this attribute to 'true'. Default value is 'false'.`,
			},
			"reject_invalid_jars": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: `Reject the caching of jar files that are found to be invalid. For example, pseudo jars retrieved behind a "captive portal". Default value is 'false'.`,
			},
		},
		repository.RepoLayoutRefSchema(rclass, packageType),
	)
}

func UnpackBaseRemoteRepo(s *schema.ResourceData, packageType string) RepositoryRemoteBaseParams {
	d := &utilsdk.ResourceData{ResourceData: s}

	repo := RepositoryRemoteBaseParams{
		Rclass:                            "remote",
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
		ClientTlsCertificate:              d.GetString("client_tls_certificate", false),
		PriorityResolution:                d.GetBool("priority_resolution", false),
		ListRemoteFolderItems:             d.GetBool("list_remote_folder_items", false),
		MismatchingMimeTypeOverrideList:   d.GetString("mismatching_mime_types_override_list", false),
		DisableURLNormalization:           d.GetBool("disable_url_normalization", false),
	}
	if v, ok := d.GetOk("content_synchronisation"); ok {
		contentSynchronisationConfig := v.([]interface{})[0].(map[string]interface{})
		enabled := contentSynchronisationConfig["enabled"].(bool)
		statisticsEnabled := contentSynchronisationConfig["statistics_enabled"].(bool)
		propertiesEnabled := contentSynchronisationConfig["properties_enabled"].(bool)
		sourceOriginAbsenceDetection := contentSynchronisationConfig["source_origin_absence_detection"].(bool)
		repo.ContentSynchronisation = &repository.ContentSynchronisation{
			Enabled: enabled,
			Statistics: repository.ContentSynchronisationStatistics{
				Enabled: statisticsEnabled,
			},
			Properties: repository.ContentSynchronisationProperties{
				Enabled: propertiesEnabled,
			},
			Source: repository.ContentSynchronisationSource{
				OriginAbsenceDetection: sourceOriginAbsenceDetection,
			},
		}
	}
	return repo
}

func UnpackVcsRemoteRepo(s *schema.ResourceData) RepositoryVcsParams {
	d := &utilsdk.ResourceData{ResourceData: s}
	return RepositoryVcsParams{
		VcsGitProvider:    d.GetString("vcs_git_provider", false),
		VcsGitDownloadUrl: d.GetString("vcs_git_download_url", false),
	}
}

func UnpackJavaRemoteRepo(s *schema.ResourceData, repoType string) JavaRemoteRepo {
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
	}
}

var resourceV1 = &schema.Resource{
	Schema: baseRemoteRepoSchemaV1,
}

var resourceV2 = &schema.Resource{
	Schema: baseRemoteRepoSchemaV2,
}

func mkResourceSchema(skeema map[string]*schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor repository.Constructor) *schema.Resource {
	var reader = repository.MkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: repository.MkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: repository.MkRepoUpdate(unpack, reader),
		DeleteContext: repository.DeleteRepo,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceV1.CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceStateUpgradeV1,
				Version: 1,
			},
			{
				// this only works because the schema hasn't changed, except the removal of default value
				// from `project_key` attribute.
				Type:    resourceV2.CoreConfigSchema().ImpliedType(),
				Upgrade: repository.ResourceUpgradeProjectKey,
				Version: 2,
			},
		},

		Schema:        skeema,
		SchemaVersion: 3,
		CustomizeDiff: customdiff.All(
			repository.ProjectEnvironmentsDiff,
			verifyExternalDependenciesDockerAndHelm,
			repository.VerifyDisableProxy,
			verifyRemoteRepoLayoutRef,
		),
	}
}

func verifyExternalDependenciesDockerAndHelm(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	// Skip the verification if schema doesn't have `external_dependencies_enabled` attribute (only docker and helm have it)
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

func verifyRemoteRepoLayoutRef(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	ref := diff.Get("remote_repo_layout_ref").(string)
	isChanged := diff.HasChange("remote_repo_layout_ref")

	if isChanged && len(ref) == 0 {
		return fmt.Errorf("empty remote_repo_layout_ref will not remove the actual attribute value and will be ignored by the API, " +
			"thus will create a state drift on the next plan. Please add the attribute, according to the repository type")
	}

	return nil
}
