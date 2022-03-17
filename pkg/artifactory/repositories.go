package artifactory

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const repositoriesEndpoint = "artifactory/api/repositories/"

type LocalRepositoryBaseParams struct {
	Key                    string   `hcl:"key" json:"key,omitempty"`
	ProjectKey             string   `json:"projectKey"`
	ProjectEnvironments    []string `json:"environments"`
	Rclass                 string   `json:"rclass"`
	PackageType            string   `hcl:"package_type" json:"packageType,omitempty"`
	Description            string   `hcl:"description" json:"description,omitempty"`
	Notes                  string   `hcl:"notes" json:"notes,omitempty"`
	IncludesPattern        string   `hcl:"includes_pattern" json:"includesPattern,omitempty"`
	ExcludesPattern        string   `hcl:"excludes_pattern" json:"excludesPattern,omitempty"`
	RepoLayoutRef          string   `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	BlackedOut             *bool    `hcl:"blacked_out" json:"blackedOut,omitempty"`
	XrayIndex              bool     `json:"xrayIndex"`
	PropertySets           []string `hcl:"property_sets" json:"propertySets,omitempty"`
	ArchiveBrowsingEnabled *bool    `hcl:"archive_browsing_enabled" json:"archiveBrowsingEnabled,omitempty"`
	DownloadRedirect       *bool    `hcl:"download_direct" json:"downloadRedirect,omitempty"`
	PriorityResolution     bool     `hcl:"priority_resolution" json:"priorityResolution"`
}

var compressionFormats = map[string]*schema.Schema{
	"index_compression_formats": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Set:      schema.HashString,
		Optional: true,
	},
}

func (bp LocalRepositoryBaseParams) Id() string {
	return bp.Key
}

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

type RemoteRepositoryBaseParams struct {
	Key                      string   `hcl:"key" json:"key,omitempty"`
	ProjectKey               string   `json:"projectKey"`
	ProjectEnvironments      []string `json:"environments"`
	Rclass                   string   `json:"rclass"`
	PackageType              string   `hcl:"package_type" json:"packageType,omitempty"`
	Url                      string   `hcl:"url" json:"url"`
	Username                 string   `hcl:"username" json:"username,omitempty"`
	Password                 string   `hcl:"password" json:"password,omitempty"`
	Proxy                    string   `hcl:"proxy" json:"proxy"`
	Description              string   `hcl:"description" json:"description,omitempty"`
	Notes                    string   `hcl:"notes" json:"notes,omitempty"`
	IncludesPattern          string   `hcl:"includes_pattern" json:"includesPattern,omitempty"`
	ExcludesPattern          string   `hcl:"excludes_pattern" json:"excludesPattern,omitempty"`
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
	RetrievalCachePeriodSecs int      `hcl:"retrieval_cache_period_seconds" json:"retrievalCachePeriodSecs,omitempty"`
	// doesn't appear in the body when calling get. Hence no HCL
	FailedRetrievalCachePeriodSecs    int                     `json:"failedRetrievalCachePeriodSecs,omitempty"`
	MissedRetrievalCachePeriodSecs    int                     `hcl:"missed_cache_period_seconds" json:"missedRetrievalCachePeriodSecs"`
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
	MismatchingMimeTypeOverrideList   string                  `hcl:"mismatching_mime_types_override_list" json:"mismatchingMimeTypesOverrideList"`
	ListRemoteFolderItems             bool                    `json:"listRemoteFolderItems"`
}

func (bp RemoteRepositoryBaseParams) Id() string {
	return bp.Key
}

type VirtualRepositoryBaseParams struct {
	Key                                           string   `hcl:"key" json:"key,omitempty"`
	ProjectKey                                    string   `json:"projectKey"`
	ProjectEnvironments                           []string `json:"environments"`
	Rclass                                        string   `json:"rclass"`
	PackageType                                   string   `hcl:"package_type" json:"packageType,omitempty"`
	Description                                   string   `hcl:"description" json:"description,omitempty"`
	Notes                                         string   `hcl:"notes" json:"notes,omitempty"`
	IncludesPattern                               string   `hcl:"includes_pattern" json:"includesPattern,omitempty"`
	ExcludesPattern                               string   `hcl:"excludes_pattern" json:"excludesPattern,omitempty"`
	RepoLayoutRef                                 string   `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	Repositories                                  []string `hcl:"repositories" json:"repositories,omitempty"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts bool     `hcl:"artifactory_requests_can_retrieve_remote_artifacts" json:"artifactoryRequestsCanRetrieveRemoteArtifacts,omitempty"`
	DefaultDeploymentRepo                         string   `hcl:"default_deployment_repo" json:"defaultDeploymentRepo,omitempty"`
}

type VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs struct {
	VirtualRepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `hcl:"retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs"`
}

func (bp VirtualRepositoryBaseParams) Id() string {
	return bp.Key
}

type ReadFunc func(d *schema.ResourceData, m interface{}) error

// Constructor Must return a pointer to a struct. When just returning a struct, resty gets confused and thinks it's a map
type Constructor func() interface{}

// UnpackFunc must return a pointer to a struct and the resource id
type UnpackFunc func(s *schema.ResourceData) (interface{}, string, error)

type PackFunc func(repo interface{}, d *schema.ResourceData) error

var retryOnMergeError = func() func(response *resty.Response, _r error) bool {
	var mergeAndSaveRegex = regexp.MustCompile(".*Could not merge and save new descriptor.*")
	return func(response *resty.Response, _r error) bool {
		return mergeAndSaveRegex.MatchString(string(response.Body()[:]))
	}
}()

func mkRepoCreate(unpack UnpackFunc, read schema.ReadContextFunc) schema.CreateContextFunc {

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().AddRetryCondition(retryOnMergeError).SetBody(repo).Put(repositoriesEndpoint + key)

		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(key)
		return read(ctx, d, m)
	}
}

func mkRepoRead(pack PackFunc, construct Constructor) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo := construct()
		// repo must be a pointer
		resp, err := m.(*resty.Client).R().SetResult(repo).Get(repositoriesEndpoint + d.Id())

		if err != nil {
			if resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.FromErr(pack(repo, d))
	}
}

func mkRepoUpdate(unpack UnpackFunc, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().AddRetryCondition(retryOnMergeError).SetBody(repo).Post(repositoriesEndpoint + d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(key)
		return read(ctx, d, m)
	}
}

func deleteRepo(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(*resty.Client).R().AddRetryCondition(retryOnMergeError).Delete(repositoriesEndpoint + d.Id())

	if err != nil && (resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound)) {
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}

var neverRetry = func(response *resty.Response, err error) bool {
	return false
}

var retry400 = func(response *resty.Response, err error) bool {
	return response.StatusCode() == 400
}

func checkRepo(id string, request *resty.Request) (*resty.Response, error) {
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return request.Head(repositoriesEndpoint + id)
}

func repoExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := checkRepo(d.Id(), m.(*resty.Client).R().AddRetryCondition(retry400))
	return err == nil, err
}

var repoTypeValidator = validation.StringInSlice(repoTypesSupported, false)

var repoKeyValidator = validation.All(
	validation.StringDoesNotMatch(regexp.MustCompile("^[0-9].*"), "repo key cannot start with a number"),
	validation.StringDoesNotContainAny(" !@#$%^&*()+={}[]:;<>,/?~`|\\"),
)

var repoTypesSupported = []string{
	"alpine",
	"bower",
	"cargo",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"debian",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"gradle",
	"helm",
	"ivy",
	"maven",
	"npm",
	"nuget",
	"opkg",
	"p2",
	"puppet",
	"pypi",
	"rpm",
	"sbt",
	"vagrant",
	"vcs",
}

var federatedRepoTypesSupported = []string{
	"alpine",
	"bower",
	"cargo",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"debian",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"gradle",
	"helm",
	"ivy",
	"maven",
	"npm",
	"nuget",
	"opkg",
	"puppet",
	"pypi",
	"rpm",
	"sbt",
	"vagrant",
}

var repoTypesLikeGeneric = []string{
	"bower",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"cran",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"helm",
	"npm",
	"opkg",
	"puppet",
	"pypi",
	"vagrant",
}

var remoteRepoTypesLikeGeneric = []string{
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
	"puppet",
	"rpm",
}

var gradleLikeRepoTypes = []string{
	"gradle",
	"sbt",
	"ivy",
}

var projectEnvironmentsSupported = []string{"DEV", "PROD"}

func repoLayoutRefSchema(repositoryType string, packageType string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"repo_layout_ref": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: getDefaultRepoLayoutRef(repositoryType, packageType),
			Description: "Repository layout key for the local repository",
		},
	}
}

var baseLocalRepoSchema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repoKeyValidator,
		Description:  "A mandatory identifier for the repository that must be unique. It cannot begin with a number or contain spaces or special characters.",
	},
	"project_key": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: projectKeyValidator,
		Description:      "Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		MinItems:    1,
		MaxItems:    2,
		Set:         schema.HashString,
		Optional:    true,
		Description: `Project environment for assigning this repository to. Allow values: "DEV" or "PROD"`,
	},
	"package_type": {
		Type:     schema.TypeString,
		Required: false,
		Computed: true,
		ForceNew: true,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"notes": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"includes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true, Description: "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"excludes_pattern": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: repoLayoutRefSchemaOverrideValidator,
		Description:      "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
	},
	"blacked_out": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, the repository does not participate in artifact resolution and new artifacts cannot be deployed.",
	},
	"xray_index": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.",
	},
	"priority_resolution": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
	},
	"property_sets": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Set:         schema.HashString,
		Optional:    true,
		Description: "List of property set name",
	},
	"archive_browsing_enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "When set, you may view content such as HTML or Javadoc files directly from Artifactory.\nThis may not be safe and therefore requires strict content moderation to prevent malicious users from uploading content that may compromise security (e.g., cross-site scripting attacks).",
	},
	"download_direct": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "When set, download requests to this repository will redirect the client to download the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.",
	},
}

var baseRemoteRepoSchema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repoKeyValidator,
	},
	"project_key": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: projectKeyValidator,
		Description:      "Project key for assigning this repository to. Must be 3 - 10 lowercase alphanumeric characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		MaxItems:    2,
		Set:         schema.HashString,
		Optional:    true,
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
	},
	"username": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"password": {
		Type:      schema.TypeString,
		Optional:  true,
		Sensitive: true,
		StateFunc: getMD5Hash,
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
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"excludes_pattern": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: repoLayoutRefSchemaOverrideValidator,
		Description:      "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
	},
	"remote_repo_layout_ref": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Repository layout key for the remote layout mapping",
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
		Optional: true, Description: "List of property set name",
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
					Description: `(Optional) If set, Remote repository proxies a local or remote repository from another instance of Artifactory. Default value is 'false'.`,
				},
				"statistics_enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: `(Optional) If set, Artifactory will notify the remote instance whenever an artifact in the Smart Remote Repository is downloaded locally so that it can update its download counter. Note that if this option is not set, there may be a discrepancy between the number of artifacts reported to have been downloaded in the different Artifactory instances of the proxy chain. Default value is 'false'.`,
				},
				"properties_enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: `(Optional) If set, properties for artifacts that have been cached in this repository will be updated if they are modified in the artifact hosted at the remote Artifactory instance. The trigger to synchronize the properties is download of the artifact from the remote repository cache of the local Artifactory instance. Default value is 'false'.`,
				},
				"source_origin_absence_detection": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: `(Optional) If set, Artifactory displays an indication on cached items if they have been deleted from the corresponding repository in the remote Artifactory instance. Default value is 'false'`,
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
		Description: `(Optional) Lists the items of remote folders in simple and list browsing. The remote content is cached according to the value of the 'Retrieval Cache Period'. Default value is 'false'.`,
	},
	"mismatching_mime_types_override_list": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: commaSeperatedList,
		StateFunc: func(thing interface{}) string {
			fields := strings.Fields(thing.(string))
			sort.Strings(fields)
			return strings.Join(fields, ",")
		},
	},
}

var baseVirtualRepoSchema = map[string]*schema.Schema{
	"key": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The Repository Key. A mandatory identifier for the repository and must be unique. It cannot begin with a number or contain spaces or special characters. For local repositories, we recommend using a '-local' suffix (e.g. 'libs-release-local').",
	},
	"project_key": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: projectKeyValidator,
		Description:      "Project key for assigning this repository to. Must be 3 - 10 lowercase alphanumeric characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		MaxItems:    2,
		Set:         schema.HashString,
		Optional:    true,
		Description: `Project environment for assigning this repository to. Allow values: "DEV" or "PROD"`,
	},
	"package_type": {
		Type:        schema.TypeString,
		Required:    false,
		Computed:    true,
		ForceNew:    true,
		Description: "The Package Type. This must be specified when the repository is created, and once set, cannot be changed.",
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "A free text field that describes the content and purpose of the repository.\nIf you choose to insert a link into this field, clicking the link will prompt the user to confirm that they might be redirected to a new domain.",
	},
	"notes": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "A free text field to add additional notes about the repository. These are only visible to the administrator.",
	},
	"includes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Default:  "**/*",
		Description: "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. " +
			"When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*." +
			"By default no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: repoLayoutRefSchemaOverrideValidator,
		Description:      "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
	},
	"repositories": {
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "The effective list of actual repositories included in this virtual repository.",
	},

	"artifactory_requests_can_retrieve_remote_artifacts": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Whether the virtual repository should search through remote repositories when trying to resolve an artifact requested by another Artifactory instance.",
	},
	"default_deployment_repo": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Default repository to deploy artifacts.",
	},
	"retrieval_cache_period_seconds": {
		Type:         schema.TypeInt,
		Optional:     true,
		Default:      7200,
		Description:  "This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.",
		ValidateFunc: validation.IntAtLeast(0),
	},
}

var baseFederatedRepoSchema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repoKeyValidator,
	},
	"project_key": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: projectKeyValidator,
		Description:      "Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		MinItems:    1,
		MaxItems:    2,
		Set:         schema.HashString,
		Optional:    true,
		Description: `Project environment for assigning this repository to. Allow values: "DEV" or "PROD"`,
	},
	"package_type": {
		Type:     schema.TypeString,
		Required: false,
		Computed: true,
		ForceNew: true,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
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
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: repoLayoutRefSchemaOverrideValidator,
		Description:      "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
	},
	"blacked_out": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
	},
	"xray_index": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"priority_resolution": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
	},
	"property_sets": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
		Optional: true,
	},
	"archive_browsing_enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "When set, you may view content such as HTML or Javadoc files directly from Artifactory.\nThis may not be safe and therefore requires strict content moderation to prevent malicious users from uploading content that may compromise security (e.g., cross-site scripting attacks).",
	},
	"download_direct": {
		Type:     schema.TypeBool,
		Optional: true,
	},
}

func unpackBaseRepo(rclassType string, s *schema.ResourceData, packageType string) LocalRepositoryBaseParams {
	d := &ResourceData{s}
	return LocalRepositoryBaseParams{
		Rclass:                 rclassType,
		Key:                    d.getString("key", false),
		ProjectKey:             d.getString("project_key", false),
		ProjectEnvironments:    d.getSet("project_environments"),
		PackageType:            packageType,
		Description:            d.getString("description", false),
		Notes:                  d.getString("notes", false),
		IncludesPattern:        d.getString("includes_pattern", false),
		ExcludesPattern:        d.getString("excludes_pattern", false),
		RepoLayoutRef:          d.getString("repo_layout_ref", false),
		BlackedOut:             d.getBoolRef("blacked_out", false),
		ArchiveBrowsingEnabled: d.getBoolRef("archive_browsing_enabled", false),
		PropertySets:           d.getSet("property_sets"),
		XrayIndex:              d.getBool("xray_index", false),
		DownloadRedirect:       d.getBoolRef("download_direct", false),
		PriorityResolution:     d.getBool("priority_resolution", false),
	}
}

func unpackBaseRemoteRepo(s *schema.ResourceData, packageType string) RemoteRepositoryBaseParams {
	d := &ResourceData{s}

	repo := RemoteRepositoryBaseParams{
		Rclass:                   "remote",
		Key:                      d.getString("key", false),
		ProjectKey:               d.getString("project_key", false),
		ProjectEnvironments:      d.getSet("project_environments"),
		PackageType:              packageType, // must be set independently
		Url:                      d.getString("url", false),
		Username:                 d.getString("username", true),
		Password:                 d.getString("password", true),
		Proxy:                    d.getString("proxy", true),
		Description:              d.getString("description", true),
		Notes:                    d.getString("notes", true),
		IncludesPattern:          d.getString("includes_pattern", true),
		ExcludesPattern:          d.getString("excludes_pattern", true),
		RepoLayoutRef:            d.getString("repo_layout_ref", true),
		HardFail:                 d.getBoolRef("hard_fail", true),
		Offline:                  d.getBoolRef("offline", true),
		BlackedOut:               d.getBoolRef("blacked_out", true),
		XrayIndex:                d.getBool("xray_index", true),
		StoreArtifactsLocally:    d.getBoolRef("store_artifacts_locally", true),
		SocketTimeoutMillis:      d.getInt("socket_timeout_millis", true),
		LocalAddress:             d.getString("local_address", true),
		RetrievalCachePeriodSecs: d.getInt("retrieval_cache_period_seconds", true),
		// Not returned in the GET
		//FailedRetrievalCachePeriodSecs:    d.getInt("failed_retrieval_cache_period_secs", true),
		MissedRetrievalCachePeriodSecs:    d.getInt("missed_cache_period_seconds", false),
		UnusedArtifactsCleanupEnabled:     d.getBoolRef("unused_artifacts_cleanup_period_enabled", true),
		UnusedArtifactsCleanupPeriodHours: d.getInt("unused_artifacts_cleanup_period_hours", true),
		AssumedOfflinePeriodSecs:          d.getInt("assumed_offline_period_secs", true),
		ShareConfiguration:                d.getBoolRef("share_configuration", true),
		SynchronizeProperties:             d.getBoolRef("synchronize_properties", true),
		BlockMismatchingMimeTypes:         d.getBoolRef("block_mismatching_mime_types", true),
		PropertySets:                      d.getSet("property_sets"),
		AllowAnyHostAuth:                  d.getBoolRef("allow_any_host_auth", true),
		EnableCookieManagement:            d.getBoolRef("enable_cookie_management", true),
		BypassHeadRequests:                d.getBoolRef("bypass_head_requests", true),
		ClientTlsCertificate:              d.getString("client_tls_certificate", true),
		PriorityResolution:                d.getBool("priority_resolution", false),
		ListRemoteFolderItems:             d.getBool("list_remote_folder_items", false),
		MismatchingMimeTypeOverrideList:   d.getString("mismatching_mime_types_override_list", false),
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

// Special handling for field that requires non-existant value for RT
//
// Artifactory REST API will not accept empty string or null to reset value to not set
// Instead, using a non-existant value works as a workaround
// To ensure we don't accidentally set the value to a valid value, we use a UUID v4 string
func handleResetWithNonExistantValue(d *ResourceData, key string) string {
	value := d.getString(key, false)

	// When value has changed and is empty string, then it has been removed from
	// the Terraform configuration.
	if value == "" && d.HasChange(key) {
		return fmt.Sprintf("non-existant-value-%d", randomInt())
	}

	return value
}

func unpackBaseVirtRepo(s *schema.ResourceData, packageType string) VirtualRepositoryBaseParams {
	d := &ResourceData{s}

	return VirtualRepositoryBaseParams{
		Key:                 d.getString("key", false),
		Rclass:              "virtual",
		ProjectKey:          d.getString("project_key", false),
		ProjectEnvironments: d.getSet("project_environments"),
		PackageType:         packageType, // must be set independently
		IncludesPattern:     d.getString("includes_pattern", false),
		ExcludesPattern:     d.getString("excludes_pattern", false),
		RepoLayoutRef:       d.getString("repo_layout_ref", false),
		ArtifactoryRequestsCanRetrieveRemoteArtifacts: d.getBool("artifactory_requests_can_retrieve_remote_artifacts", false),
		Repositories:          d.getList("repositories"),
		Description:           d.getString("description", false),
		Notes:                 d.getString("notes", false),
		DefaultDeploymentRepo: handleResetWithNonExistantValue(d, "default_deployment_repo"),
	}
}

func unpackBaseVirtRepoWithRetrievalCachePeriodSecs(s *schema.ResourceData, packageType string) VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs {
	d := &ResourceData{s}

	return VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs{
		VirtualRepositoryBaseParams:     unpackBaseVirtRepo(s, packageType),
		VirtualRetrievalCachePeriodSecs: d.getInt("retrieval_cache_period_seconds", false),
	}
}

// universalUnpack - todo implement me
func universalUnpack(payload reflect.Type, s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	var t = reflect.TypeOf(payload)
	var v = reflect.ValueOf(payload)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	//lookup := map[reflect.Kind]func(field, val reflect.Value) {
	//	reflect.String: func(field, val reflect.Value)  {
	//		val.SetString(field.String())
	//	},
	//}
	for i := 0; i < t.NumField(); i++ {
		thing := v.Field(i)

		switch thing.Kind() {
		case reflect.String:
			v.SetString(thing.String())
		case reflect.Int:
			v.SetInt(thing.Int())
		case reflect.Bool:
			v.SetBool(thing.Bool())
		}
	}
	result := KeyPairPayLoad{
		PairName:    d.getString("pair_name", false),
		PairType:    d.getString("pair_type", false),
		Alias:       d.getString("alias", false),
		PrivateKey:  strings.ReplaceAll(d.getString("private_key", false), "\t", ""),
		PublicKey:   strings.ReplaceAll(d.getString("public_key", false), "\t", ""),
		Unavailable: d.getBool("unavailable", false),
	}
	return &result, result.PairName, nil
}

type AutoMapper func(field reflect.StructField, thing reflect.Value) map[string]interface{}

func checkForHcl(mapper AutoMapper) AutoMapper {
	return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
		if field.Tag.Get("hcl") != "" {
			return mapper(field, thing)
		}
		return map[string]interface{}{}
	}
}
func findInspector(kind reflect.Kind) AutoMapper {
	switch kind {
	case reflect.Struct:
		return func(f reflect.StructField, t reflect.Value) map[string]interface{} {
			return lookup(t.Interface(), nil)
		}
	case reflect.Ptr:
		return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
			deref := reflect.Indirect(thing)
			if deref.CanAddr() {
				result := deref.Interface()
				if deref.Kind() == reflect.Struct {
					result = []interface{}{lookup(deref.Interface(), nil)}
				}
				return map[string]interface{}{
					fieldToHcl(field): result,
				}
			}
			return map[string]interface{}{}
		}
	case reflect.Slice:
		return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
			return map[string]interface{}{
				fieldToHcl(field): castToInterfaceArr(thing.Interface().([]string)),
			}
		}
	}
	return func(field reflect.StructField, thing reflect.Value) map[string]interface{} {
		return map[string]interface{}{
			fieldToHcl(field): thing.Interface(),
		}
	}
}

// fieldToHcl this function is meant to use the HCL provided in the tag, or create a snake_case from the field name
// it actually works as expected, but dynamically working with these names was catching edge cases everywhere and
// it was/is a time sink to catch.
func fieldToHcl(field reflect.StructField) string {

	if field.Tag.Get("hcl") != "" {
		return field.Tag.Get("hcl")
	}
	var lowerFields []string
	rgx := regexp.MustCompile("([A-Z][a-z]+)")
	fields := rgx.FindAllStringSubmatch(field.Name, -1)
	for _, matches := range fields {
		for _, match := range matches[1:] {
			lowerFields = append(lowerFields, strings.ToLower(match))
		}
	}
	result := strings.Join(lowerFields, "_")
	return result
}

func lookup(payload interface{}, predicate HclPredicate) map[string]interface{} {

	if predicate == nil {
		predicate = allowAllPredicate
	}

	values := map[string]interface{}{}
	var t = reflect.TypeOf(payload)
	var v = reflect.ValueOf(payload)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		thing := v.Field(i)

		shouldLookup := true
		if thing.Kind() != reflect.Struct {
			hcl := fieldToHcl(field)
			shouldLookup = predicate(hcl)
		}

		if shouldLookup {
			typeInspector := findInspector(thing.Kind())
			for key, value := range typeInspector(field, thing) {
				if _, ok := values[key]; !ok {
					values[key] = value
				}
			}
		}
	}
	return values
}

func anyHclPredicate(predicates ...HclPredicate) HclPredicate {
	return func(hcl string) bool {
		for _, predicate := range predicates {
			if predicate(hcl) {
				return true
			}
		}
		return false
	}
}

func allHclPredicate(predicates ...HclPredicate) HclPredicate {
	return func(hcl string) bool {
		for _, predicate := range predicates {
			if !predicate(hcl) {
				return false
			}
		}
		return true
	}
}

var noClass = ignoreHclPredicate("class", "rclass")

var allowAllPredicate = func(hcl string) bool {
	return true
}

func ignoreHclPredicate(names ...string) HclPredicate {
	set := map[string]interface{}{}
	for _, name := range names {
		set[name] = nil
	}
	return func(hcl string) bool {
		_, found := set[hcl]
		return !found
	}
}

func composePacker(packers ...PackFunc) PackFunc {
	return func(repo interface{}, d *schema.ResourceData) error {
		var errors []error

		for _, packer := range packers {
			err := packer(repo, d)
			if err != nil {
				errors = append(errors, err)
			}
		}
		if errors != nil && len(errors) > 0 {
			return fmt.Errorf("failed saving state %q", errors)
		}
		return nil
	}
}

var defaultPacker = universalPack(noClass)

func inSchema(skeema map[string]*schema.Schema) func(payload interface{}, d *schema.ResourceData) error {
	return universalPack(schemaHasKey(skeema))
}

// universalPack consider making this a function that takes a predicate of what to include and returns
// a function that does the job. This would allow for the legacy code to specify which keys to keep and not
func universalPack(predicate HclPredicate) func(payload interface{}, d *schema.ResourceData) error {

	return func(payload interface{}, d *schema.ResourceData) error {
		setValue := mkLens(d)

		var errors []error

		values := lookup(payload, predicate)

		for hcl, value := range values {
			if predicate != nil && predicate(hcl) {
				errors = setValue(hcl, value)
			}
		}

		if errors != nil && len(errors) > 0 {
			return fmt.Errorf("failed saving state %q", errors)
		}
		return nil
	}
}

func projectEnvironmentsDiff(_ context.Context, diff *schema.ResourceDiff, i interface{}) error {
	if data, ok := diff.GetOk("project_environments"); ok {
		projectEnvironments := data.(*schema.Set).List()

		for _, projectEnvironment := range projectEnvironments {
			if !contains(projectEnvironmentsSupported, projectEnvironment.(string)) {
				return fmt.Errorf("project_environment %s not allowed", projectEnvironment)
			}
		}
	}

	return nil
}

func mkResourceSchema(skeema map[string]*schema.Schema, packer PackFunc, unpack UnpackFunc, constructor Constructor) *schema.Resource {
	var reader = mkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: mkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: mkRepoUpdate(unpack, reader),
		DeleteContext: deleteRepo,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        skeema,
		CustomizeDiff: projectEnvironmentsDiff,
	}
}

//Returns random string from a map[string]string
func selectRandomFromMapOfStrings(m map[string]string) string {
	mapLength := len(m)
	allValues := make([]string, 0, mapLength)
	for _, value := range m {
		allValues = append(allValues, value)
	}
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return allValues[r1.Intn(mapLength)]
}

func isSelectRandom(opts ...bool) bool {
	selectRandomFlag := false
	for i, val := range opts {
		switch i {
		case 0:
			selectRandomFlag = val
		default:
			fmt.Printf("Option index is not defined. Index: %v, value: %v\n", i, val)
		}
	}
	return selectRandomFlag
}

type SupportedRepoClasses struct {
	RepoLayoutRef      string
	SupportedRepoTypes map[string]bool
}

//Consolidated list of Default Repo Layout for all Package Types with active Repo Types
var defaultRepoLayoutMap = map[string]SupportedRepoClasses{
	"alpine":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"bower":     {RepoLayoutRef: "bower-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cran":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cargo":     {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "federated": true}},
	"chef":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cocoapods": {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "federated": true}},
	"composer":  {RepoLayoutRef: "composer-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"conan":     {RepoLayoutRef: "conan-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"conda":     {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"debian":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"docker":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gems":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"generic":   {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gitlfs":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"go":        {RepoLayoutRef: "go-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gradle":    {RepoLayoutRef: "maven-2-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"helm":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"ivy":       {RepoLayoutRef: "ivy-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"maven":     {RepoLayoutRef: "maven-2-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"npm":       {RepoLayoutRef: "npm-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"nuget":     {RepoLayoutRef: "nuget-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"opkg":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"p2":        {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"remote": true, "virtual": true}},
	"pub":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"puppet":    {RepoLayoutRef: "puppet-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"pypi":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"sbt":       {RepoLayoutRef: "sbt-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"vagrant":   {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "federated": true}},
	"vcs":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"remote": true}},
	"rpm":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
}

//Return the default repo layout by Repository Type & Package Type
func getDefaultRepoLayoutRef(repositoryType string, packageType string) func() (interface{}, error) {
	return func() (interface{}, error) {
		if v, ok := defaultRepoLayoutMap[packageType].SupportedRepoTypes[repositoryType]; ok && v {
			return defaultRepoLayoutMap[packageType].RepoLayoutRef, nil
		}
		return "", fmt.Errorf("default repo layout not found for repository type %v & package type %v", repositoryType, packageType)
	}
}

type Identifiable interface {
	Id() string
}
