package remote

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/validators"
)
type ContentSynchronisation struct {
	Enabled bool `hcl:"enabled" json:"enables,omitempty"`
}
type RepositoryBaseParams struct {
	Key                      string `hcl:"key" json:"key,omitempty"`
	Rclass                   string `json:"rclass"`
	PackageType              string `hcl:"package_type" json:"packageType,omitempty"`
	Url                      string `hcl:"url" json:"url"`
	Username                 string `hcl:"username" json:"username,omitempty"`
	Password                 string `hcl:"password" json:"password,omitempty"`
	Proxy                    string `hcl:"proxy" json:"proxy"`
	Description              string `hcl:"description" json:"description,omitempty"`
	Notes                    string `hcl:"notes" json:"notes,omitempty"`
	IncludesPattern          string `hcl:"includes_pattern" json:"includesPattern,omitempty"`
	ExcludesPattern          string `hcl:"excludes_pattern" json:"excludesPattern,omitempty"`
	RepoLayoutRef            string `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	HardFail                 *bool  `hcl:"hard_fail" json:"hardFail,omitempty"`
	Offline                  *bool  `hcl:"offline" json:"offline,omitempty"`
	BlackedOut               *bool  `hcl:"blacked_out" json:"blackedOut,omitempty"`
	XrayIndex                *bool  `hcl:"xray_index" json:"xrayIndex,omitempty"`
	PropagateQueryParams     bool   `hcl:"propagate_query_params" json:"propagateQueryParams"`
	PriorityResolution       bool   `hcl:"priority_resolution" json:"priorityResolution"`
	StoreArtifactsLocally    *bool  `hcl:"store_artifacts_locally" json:"storeArtifactsLocally,omitempty"`
	SocketTimeoutMillis      int    `hcl:"socket_timeout_millis" json:"socketTimeoutMillis,omitempty"`
	LocalAddress             string `hcl:"local_address" json:"localAddress,omitempty"`
	RetrievalCachePeriodSecs int    `hcl:"retrieval_cache_period_seconds" json:"retrievalCachePeriodSecs,omitempty"`
	// doesn't appear in the body when calling get. Hence no HCL
	FailedRetrievalCachePeriodSecs    int                     `json:"failedRetrievalCachePeriodSecs,omitempty"`
	MissedRetrievalCachePeriodSecs    int                     `hcl:"missed_cache_period_seconds" json:"missedRetrievalCachePeriodSecs,omitempty"`
	UnusedArtifactsCleanupEnabled     *bool                   `hcl:"unused_artifacts_cleanup_period_enabled" json:"unusedArtifactsCleanupEnabled,omitempty"`
	UnusedArtifactsCleanupPeriodHours int                     `hcl:"unused_artifacts_cleanup_period_hours" json:"unusedArtifactsCleanupPeriodHours,omitempty"`
	AssumedOfflinePeriodSecs          int                     `hcl:"assumed_offline_period_secs" json:"assumedOfflinePeriodSecs,omitempty"`
	ShareConfiguration                *bool                   `hcl:"share_configuration" json:"shareConfiguration,omitempty"`
	SynchronizeProperties             *bool                   `hcl:"synchronize_properties" json:"synchronizeProperties,omitempty"`
	BlockMismatchingMimeTypes         *bool                   `hcl:"block_mismatching_mime_types" json:"blockMismatchingMimeTypes,omitempty"`
	PropertySets                      []string                `hcl:"property_sets" json:"propertySets,omitempty"`
	AllowAnyHostAuth                  *bool                   `hcl:"allow_any_host_auth" json:"allowAnyHostAuth,omitempty"`
	EnableCookieManagement            *bool                   `hcl:"enable_cookie_management" json:"enableCookieManagement,omitempty"`
	BypassHeadRequests                *bool                   `hcl:"bypass_head_requests" json:"bypassHeadRequests,omitempty"`
	ClientTlsCertificate              string                  `hcl:"client_tls_certificate" json:"clientTlsCertificate,omitempty"`
	ContentSynchronisation            *ContentSynchronisation `hcl:"content_synchronisation" json:"contentSynchronisation,omitempty"`
}

func (bp RepositoryBaseParams) Id() string {
	return bp.Key
}
var baseRemoteSchema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validators.RepoKeyValidator,
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
	},
	"username": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"password": {
		Type:      schema.TypeString,
		Optional:  true,
		Sensitive: true,
		StateFunc: util.GetMD5Hash,
	},
	"proxy": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
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
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"repo_layout_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"hard_fail": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
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
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
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
	},
	"local_address": {
		Type:     schema.TypeString,
		Optional: true,
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
		Description:  "This is actually the missedRetrievalCachePeriodSecs in the API",
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
		ValidateFunc: validation.IntAtLeast(0),
	},
	"assumed_offline_period_secs": {
		Type:         schema.TypeInt,
		Optional:     true,
		ValidateFunc: validation.IntAtLeast(0),
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
		Optional: true,
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
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
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
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
	},
	"propagate_query_params": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
	},
}
func unpackBaseRemoteRepo(s *schema.ResourceData, packageType string) RepositoryBaseParams {
	d := &util.ResourceData{ResourceData: s}

	repo := RepositoryBaseParams{
		Rclass: "remote",
		Key:    d.GetString("key", false),
		//must be set independently
		PackageType:              packageType,
		Url:                      d.GetString("url", false),
		Username:                 d.GetString("username", true),
		Password:                 d.GetString("password", true),
		Proxy:                    d.GetString("proxy", true),
		Description:              d.GetString("description", true),
		Notes:                    d.GetString("notes", true),
		IncludesPattern:          d.GetString("includes_pattern", true),
		ExcludesPattern:          d.GetString("excludes_pattern", true),
		RepoLayoutRef:            d.GetString("repo_layout_ref", true),
		HardFail:                 d.GetBoolRef("hard_fail", true),
		Offline:                  d.GetBoolRef("offline", true),
		BlackedOut:               d.GetBoolRef("blacked_out", true),
		XrayIndex:                d.GetBoolRef("xray_index", true),
		StoreArtifactsLocally:    d.GetBoolRef("store_artifacts_locally", true),
		SocketTimeoutMillis:      d.GetInt("socket_timeout_millis", true),
		LocalAddress:             d.GetString("local_address", true),
		RetrievalCachePeriodSecs: d.GetInt("retrieval_cache_period_seconds", true),
		// Not returned in the GET
		//FailedRetrievalCachePeriodSecs:    d.GetInt("failed_retrieval_cache_period_secs", true),
		MissedRetrievalCachePeriodSecs:    d.GetInt("missed_cache_period_seconds", true),
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
	}

	if v, ok := d.GetOk("content_synchronisation"); ok {
		contentSynchronisationConfig := v.([]interface{})[0].(map[string]interface{})
		enabled := contentSynchronisationConfig["enabled"].(bool)
		repo.ContentSynchronisation = &ContentSynchronisation{
			Enabled: enabled,
		}
	}
	return repo
}
