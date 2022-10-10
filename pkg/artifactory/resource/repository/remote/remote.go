package remote

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

type RepositoryBaseParams struct {
	Key                      string   `hcl:"key" json:"key,omitempty"`
	ProjectKey               string   `json:"projectKey"`
	ProjectEnvironments      []string `json:"environments"`
	Rclass                   string   `json:"rclass"`
	PackageType              string   `hcl:"package_type" json:"packageType,omitempty"`
	Url                      string   `hcl:"url" json:"url"`
	Username                 string   `hcl:"username" json:"username,omitempty"`
	Password                 string   `json:"password"`
	Proxy                    string   `hcl:"proxy" json:"proxy"`
	Description              string   `hcl:"description" json:"description,omitempty"`
	Notes                    string   `hcl:"notes" json:"notes,omitempty"`
	IncludesPattern          string   `hcl:"includes_pattern" json:"includesPattern,omitempty"`
	ExcludesPattern          string   `json:"excludesPattern"`
	RepoLayoutRef            string   `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	RemoteRepoLayoutRef      string   `json:"remoteRepoLayoutRef"`
	HardFail                 *bool    `hcl:"hard_fail" json:"hardFail,omitempty"`
	Offline                  *bool    `hcl:"offline" json:"offline,omitempty"`
	BlackedOut               *bool    `hcl:"blacked_out" json:"blackedOut,omitempty"`
	XrayIndex                bool     `json:"xrayIndex"`
	PropagateQueryParams     bool     `hcl:"propagate_query_params" json:"propagateQueryParams"`
	PriorityResolution       bool     `hcl:"priority_resolution" json:"priorityResolution"`
	StoreArtifactsLocally    *bool    `hcl:"store_artifacts_locally" json:"storeArtifactsLocally,omitempty"`
	SocketTimeoutMillis      int      `hcl:"socket_timeout_millis" json:"socketTimeoutMillis,omitempty"`
	LocalAddress             string   `hcl:"local_address" json:"localAddress,omitempty"`
	RetrievalCachePeriodSecs int      `hcl:"retrieval_cache_period_seconds" json:"retrievalCachePeriodSecs"`
	// doesn't appear in the body when calling get. Hence no HCL
	FailedRetrievalCachePeriodSecs    int                                `json:"failedRetrievalCachePeriodSecs,omitempty"`
	MissedRetrievalCachePeriodSecs    int                                `hcl:"missed_cache_period_seconds" json:"missedRetrievalCachePeriodSecs"`
	UnusedArtifactsCleanupEnabled     *bool                              `hcl:"unused_artifacts_cleanup_period_enabled" json:"unusedArtifactsCleanupEnabled,omitempty"`
	UnusedArtifactsCleanupPeriodHours int                                `hcl:"unused_artifacts_cleanup_period_hours" json:"unusedArtifactsCleanupPeriodHours,omitempty"`
	AssumedOfflinePeriodSecs          int                                `hcl:"assumed_offline_period_secs" json:"assumedOfflinePeriodSecs,omitempty"`
	ShareConfiguration                *bool                              `hcl:"share_configuration" json:"shareConfiguration,omitempty"`
	SynchronizeProperties             *bool                              `hcl:"synchronize_properties" json:"synchronizeProperties,omitempty"`
	BlockMismatchingMimeTypes         *bool                              `hcl:"block_mismatching_mime_types" json:"blockMismatchingMimeTypes,omitempty"`
	PropertySets                      []string                           `hcl:"property_sets" json:"propertySets,omitempty"`
	AllowAnyHostAuth                  *bool                              `hcl:"allow_any_host_auth" json:"allowAnyHostAuth,omitempty"`
	EnableCookieManagement            *bool                              `hcl:"enable_cookie_management" json:"enableCookieManagement,omitempty"`
	BypassHeadRequests                *bool                              `hcl:"bypass_head_requests" json:"bypassHeadRequests,omitempty"`
	ClientTlsCertificate              string                             `hcl:"client_tls_certificate" json:"clientTlsCertificate,omitempty"`
	ContentSynchronisation            *repository.ContentSynchronisation `hcl:"content_synchronisation" json:"contentSynchronisation,omitempty"`
	MismatchingMimeTypeOverrideList   string                             `hcl:"mismatching_mime_types_override_list" json:"mismatchingMimeTypesOverrideList"`
	ListRemoteFolderItems             bool                               `json:"listRemoteFolderItems"`
	DownloadRedirect                  bool                               `hcl:"download_direct" json:"downloadRedirect,omitempty"`
}

type JavaRemoteRepo struct {
	RepositoryBaseParams
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

func (bp RepositoryBaseParams) Id() string {
	return bp.Key
}

var RepoTypesLikeGeneric = []string{
	"alpine",
	"chef",
	"conda",
	"conan",
	"cran",
	"debian",
	"gems",
	"generic",
	"gitlfs",
	"npm",
	"opkg",
	"p2",
	"pub",
	"puppet",
	"rpm",
	"swift",
}

var BaseRemoteRepoSchema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repository.RepoKeyValidator,
		Description:  "A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.",
	},
	"project_key": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validator.ProjectKey,
		Description:      "Project key for assigning this repository to. Must be 3 - 10 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		MaxItems:    2,
		Set:         schema.HashString,
		Optional:    true,
		Computed:    true,
		Description: `Project environment for assigning this repository to. Allow values: "DEV" or "PROD"`,
	},
	"package_type": {
		Type:     schema.TypeString,
		Required: false,
		Computed: true,
		ForceNew: true,
	},
	"url": {
		Type:         schema.TypeString,
		Required:     true,
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
	"proxy": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Proxy key from Artifactory Proxies settings",
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
			// this is literally what comes back from the server
			return old == fmt.Sprintf("%s (local file cache)", new)
		},
	},
	"notes": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"includes_pattern": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "List of comma-separated artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"excludes_pattern": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "List of comma-separated artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: repository.ValidateRepoLayoutRefSchemaOverride,
		Description:      "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
	},
	"remote_repo_layout_ref": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Repository layout key for the remote layout mapping",
		Deprecated:  "This field has currently no effect, because there is no corresponding field in the API body, and it's not returned by the GET call.",
	},
	"hard_fail": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true, Description: "When set, Artifactory will return an error to the client that causes the build to fail if there is a failure to communicate with this repository.",
	},
	"offline": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "If set, Artifactory does not try to fetch remote artifacts. Only locally-cached artifacts are retrieved.",
	},
	"blacked_out": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "(A.K.A 'Ignore Repository' on the UI) When set, the repository or its local cache do not participate in artifact resolution.",
	},
	"xray_index": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.",
	},
	"store_artifacts_locally": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "When set, the repository should store cached artifacts locally. When not set, artifacts are not stored locally, and direct repository-to-client streaming is used. This can be useful for multi-server setups over a high-speed LAN, with one Artifactory caching certain data on central storage, and streaming it directly to satellite pass-though Artifactory servers.",
	},
	"socket_timeout_millis": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
		Description:  " Network timeout (in ms) to use when establishing a connection and for unanswered requests. Timing out on a network operation is considered a retrieval failure.",
	},
	"local_address": {
		Type:     schema.TypeString,
		Optional: true, Description: "The local address to be used when creating connections. Useful for specifying the interface to use on systems with multiple network interfaces.",
	},
	"retrieval_cache_period_seconds": {
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
		Description: "The metadataRetrievalTimeoutSecs field not allowed to be bigger then retrievalCachePeriodSecs field.",
		DefaultFunc: func() (interface{}, error) {
			return 7200, nil
		},
		ValidateFunc: validation.IntAtLeast(0),
	},
	"failed_retrieval_cache_period_secs": {
		Type:     schema.TypeInt,
		Computed: true,
		Deprecated: "This field is not returned in a get payload but is offered on the UI. " +
			"It's inserted here for inclusive and informational reasons. It does not function",
	},
	"missed_cache_period_seconds": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
		Description:  "The number of seconds to cache artifact retrieval misses (artifact not found). A value of 0 indicates no caching.",
	},
	"unused_artifacts_cleanup_period_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"unused_artifacts_cleanup_period_hours": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0), Description: `The number of hours to wait before an artifact is deemed "unused" and eligible for cleanup from the repository. A value of 0 means automatic cleanup of cached artifacts is disabled.`,
	},
	"assumed_offline_period_secs": {
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      300,
		ValidateFunc: validation.IntAtLeast(0),
		Description:  "The number of seconds the repository stays in assumed offline state after a connection error. At the end of this time, an online check is attempted in order to reset the offline status. A value of 0 means the repository is never assumed offline. Default to 300.",
	},
	"share_configuration": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"synchronize_properties": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "When set, remote artifacts are fetched along with their properties.",
	},
	"block_mismatching_mime_types": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.",
	},
	"property_sets": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
		Optional: true, Description: "List of property set names",
	},
	"allow_any_host_auth": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Also known as 'Lenient Host Authentication', Allow credentials of this repository to be used on requests redirected to any other host.",
	},
	"enable_cookie_management": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Enables cookie management if the remote repository uses cookies to manage client state.",
	},
	"bypass_head_requests": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Before caching an artifact, Artifactory first sends a HEAD request to the remote resource. In some remote resources, HEAD requests are disallowed and therefore rejected, even though downloading the artifact is allowed. When checked, Artifactory will bypass the HEAD request and cache the artifact directly using a GET request.",
	},
	"priority_resolution": {
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
		Description: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
	},
	"client_tls_certificate": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
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
					Description: `If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.`,
				},
				"statistics_enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: `If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.`,
				},
				"properties_enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: `If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.`,
				},
				"source_origin_absence_detection": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: `If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'`,
				},
			},
		},
	},
	"propagate_query_params": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
	},
	"list_remote_folder_items": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: `Lists the items of remote folders in simple and list browsing. The remote content is cached according to the value of the 'Retrieval Cache Period'. Default value is 'false'.`,
	},
	"mismatching_mime_types_override_list": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validator.CommaSeperatedList,
		StateFunc:        util.FormatCommaSeparatedString,
		Description:      `The set of mime types that should override the block_mismatching_mime_types setting. Eg: "application/json,application/xml". Default value is empty.`,
	},
	"download_direct": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, download requests to this repository will redirect the client to download the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only. Default value is 'false'.",
	},
}

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
		ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringIsNotEmpty, validation.IsURLWithHTTPorHTTPS)),
		Description:      `This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.`,
	},
}

func getJavaRemoteSchema(repoType string, suppressPom bool) map[string]*schema.Schema {
	return util.MergeMaps(
		BaseRemoteRepoSchema,
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
		repository.RepoLayoutRefSchema("remote", repoType),
	)
}

func UnpackBaseRemoteRepo(s *schema.ResourceData, packageType string) RepositoryBaseParams {
	d := &util.ResourceData{ResourceData: s}

	repo := RepositoryBaseParams{
		Rclass:                   "remote",
		Key:                      d.GetString("key", false),
		ProjectKey:               d.GetString("project_key", false),
		ProjectEnvironments:      d.GetSet("project_environments"),
		PackageType:              packageType, // must be set independently
		Url:                      d.GetString("url", false),
		Username:                 d.GetString("username", true),
		Password:                 d.GetString("password", false),
		Proxy:                    d.GetString("proxy", false),
		Description:              d.GetString("description", true),
		Notes:                    d.GetString("notes", true),
		IncludesPattern:          d.GetString("includes_pattern", true),
		ExcludesPattern:          d.GetString("excludes_pattern", false),
		RepoLayoutRef:            d.GetString("repo_layout_ref", true),
		HardFail:                 d.GetBoolRef("hard_fail", true),
		Offline:                  d.GetBoolRef("offline", true),
		BlackedOut:               d.GetBoolRef("blacked_out", true),
		XrayIndex:                d.GetBool("xray_index", true),
		DownloadRedirect:         d.GetBool("download_direct", false),
		PropagateQueryParams:     d.GetBool("propagate_query_params", true),
		StoreArtifactsLocally:    d.GetBoolRef("store_artifacts_locally", true),
		SocketTimeoutMillis:      d.GetInt("socket_timeout_millis", true),
		LocalAddress:             d.GetString("local_address", true),
		RetrievalCachePeriodSecs: d.GetInt("retrieval_cache_period_seconds", false),
		// Not returned in the GET
		//FailedRetrievalCachePeriodSecs:    d.GetInt("failed_retrieval_cache_period_secs", true),
		MissedRetrievalCachePeriodSecs:    d.GetInt("missed_cache_period_seconds", false),
		UnusedArtifactsCleanupEnabled:     d.GetBoolRef("unused_artifacts_cleanup_period_enabled", true),
		UnusedArtifactsCleanupPeriodHours: d.GetInt("unused_artifacts_cleanup_period_hours", true),
		AssumedOfflinePeriodSecs:          d.GetInt("assumed_offline_period_secs", true),
		ShareConfiguration:                d.GetBoolRef("share_configuration", true),
		SynchronizeProperties:             d.GetBoolRef("synchronize_properties", true),
		BlockMismatchingMimeTypes:         d.GetBoolRef("block_mismatching_mime_types", true),
		PropertySets:                      d.GetSet("property_sets"),
		AllowAnyHostAuth:                  d.GetBoolRef("allow_any_host_auth", true),
		EnableCookieManagement:            d.GetBoolRef("enable_cookie_management", true),
		BypassHeadRequests:                d.GetBoolRef("bypass_head_requests", true),
		ClientTlsCertificate:              d.GetString("client_tls_certificate", true),
		PriorityResolution:                d.GetBool("priority_resolution", false),
		ListRemoteFolderItems:             d.GetBool("list_remote_folder_items", false),
		MismatchingMimeTypeOverrideList:   d.GetString("mismatching_mime_types_override_list", false),
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
	d := &util.ResourceData{ResourceData: s}
	return RepositoryVcsParams{
		VcsGitProvider:    d.GetString("vcs_git_provider", false),
		VcsGitDownloadUrl: d.GetString("vcs_git_download_url", false),
	}
}

func UnpackJavaRemoteRepo(s *schema.ResourceData, repoType string) JavaRemoteRepo {
	d := &util.ResourceData{ResourceData: s}
	return JavaRemoteRepo{
		RepositoryBaseParams:         UnpackBaseRemoteRepo(s, repoType),
		FetchJarsEagerly:             d.GetBool("fetch_jars_eagerly", false),
		FetchSourcesEagerly:          d.GetBool("fetch_sources_eagerly", false),
		RemoteRepoChecksumPolicyType: d.GetString("remote_repo_checksum_policy_type", false),
		HandleReleases:               d.GetBool("handle_releases", false),
		HandleSnapshots:              d.GetBool("handle_snapshots", false),
		SuppressPomConsistencyChecks: d.GetBool("suppress_pom_consistency_checks", false),
		RejectInvalidJars:            d.GetBool("reject_invalid_jars", false),
	}
}
