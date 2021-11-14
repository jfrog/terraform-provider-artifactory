package remote

import (
	"fmt"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/validators"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type BowerRemoteRepositoryParams struct {
	RepositoryBaseParams
	BowerRegistryUrl string `json:"bowerRegistryUrl,omitempty"`
}

// CommonMavenGradleRemoteRepository move this to maven dedicated remote
type CommonMavenGradleRemoteRepository struct {
	FetchJarsEagerly             *bool  `hcl:"fetch_jars_eagerly" json:"fetchJarsEagerly,omitempty"`
	FetchSourcesEagerly          *bool  `hcl:"fetch_sources_eagerly" json:"fetchSourcesEagerly,omitempty"`
	RemoteRepoChecksumPolicyType string `hcl:"remote_repo_checksum_policy_type" json:"remoteRepoChecksumPolicyType,omitempty"`
	ListRemoteFolderItems        *bool  `hcl:"list_remote_folder_items" json:"listRemoteFolderItems,omitempty"`
	HandleReleases               *bool  `hcl:"handle_releases" json:"handleReleases,omitempty"`
	HandleSnapshots              *bool  `hcl:"handle_snapshots" json:"handleSnapshots,omitempty"`
	SuppressPomConsistencyChecks *bool  `hcl:"suppress_pom_consistency_checks" json:"suppressPomConsistencyChecks,omitempty"`
	RejectInvalidJars            *bool  `hcl:"reject_invalid_jars" json:"rejectInvalidJars,omitempty"`
}
type VcsRemoteRepositoryParams struct {
	RepositoryBaseParams
	VcsGitProvider        string `hcl:"vcs_git_provider" json:"vcsGitProvider,omitempty"`
	VcsType               string `hcl:"vcs_type" json:"vcsType,omitempty"`
	MaxUniqueSnapshots    int    `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots,omitempty"`
	VcsGitDownloadUrl     string `hcl:"vcs_git_download_url" json:"vcsGitDownloadUrl,omitempty"`
	ListRemoteFolderItems *bool  `hcl:"list_remote_folder_items" json:"listRemoteFolderItems,omitempty"`
}
type PypiRemoteRepositoryParams struct {
	RepositoryBaseParams
	ListRemoteFolderItems *bool  `hcl:"list_remote_folder_items" json:"listRemoteFolderItems,omitempty"`
	PypiRegistryUrl       string `hcl:"pypi_registry_url" json:"pypiRegistryUrl,omitempty"`
}
type NugetRemoteRepositoryParams struct {
	RepositoryBaseParams
	FeedContextPath          string `hcl:"feed_context_path" json:"feedContextPath,omitempty"`
	DownloadContextPath      string `hcl:"download_context_path" json:"downloadContextPath,omitempty"`
	V3FeedUrl                string `hcl:"v3_feed_url" json:"v3FeedUrl,omitempty"`
	ForceNugetAuthentication *bool  `hcl:"force_nuget_authentication" json:"forceNugetAuthentication,omitempty"`
}
type MessyRemoteRepo struct {
	RepositoryBaseParams
	BowerRemoteRepositoryParams
	CommonMavenGradleRemoteRepository
	DockerRemoteRepository
	VcsRemoteRepositoryParams
	PypiRemoteRepositoryParams
	NugetRemoteRepositoryParams
	PropagateQueryParams bool `hcl:"propagate_query_params" json:"propagateQueryParams"`
}

func (mr MessyRemoteRepo) Id() string {
	return mr.Key
}

var legacyRemoteSchema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validators.RepoKeyValidator,
	},
	"package_type": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Default:      "generic",
		ValidateFunc: validators.RepoTypeValidator,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
		DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
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
	"handle_releases": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"handle_snapshots": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"max_unique_snapshots": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
	},
	"suppress_pom_consistency_checks": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
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
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
		StateFunc:   util.GetMD5Hash,
		Description: "This field can only be used if encryption has been turned off",
	},
	"proxy": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"remote_repo_checksum_policy_type": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		ValidateFunc: validation.StringInSlice([]string{
			"generate-if-absent",
			"fail",
			"ignore-and-generate",
			"pass-thru",
		}, false),
	},
	"hard_fail": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"offline": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"blacked_out": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"store_artifacts_locally": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
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
	"missed_cache_period_seconds": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
	},
	"unused_artifacts_cleanup_period_hours": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
	},
	"fetch_jars_eagerly": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"fetch_sources_eagerly": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"share_configuration": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"synchronize_properties": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"block_mismatching_mime_types": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"property_sets": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
		Optional: true,
	},
	"allow_any_host_auth": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"enable_cookie_management": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"client_tls_certificate": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"pypi_registry_url": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"bower_registry_url": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"bypass_head_requests": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"enable_token_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"xray_index": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"vcs_type": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"vcs_git_provider": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"vcs_git_download_url": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"feed_context_path": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"download_context_path": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"v3_feed_url": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"force_nuget_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	//"metadataRetrievalTimeoutSecs": {
	//	Type: schema.TypeInt,
	//	Optional: true,
	//	Computed: true,
	//	Description: "The metadataRetrievalTimeoutSecs field not allowed to be bigger then retrievalCachePeriodSecs field.",
	//	DefaultFunc: func() (interface{}, error) {
	//		return 60, nil
	//	},
	//},
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
		Computed: true,
		DefaultFunc: func() (interface{}, error) {
			return false, nil
		},
	},
}

func ResourceArtifactoryRemoteRepository() *schema.Resource {
	// the universal pack function cannot be used because fields in the combined set of structs don't
	// appear in the HCL, such as 'Invalid address to set: []string{"external_dependencies_patterns"}' which is a docker field
	return repos.MkResourceSchema(legacyRemoteSchema, packLegacyRemoteRepo, unpackLegacyRemoteRepo, func() interface{} {
		return &MessyRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass: "remote",
			},
		}
	})
}

func unpackLegacyRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {

	d := &util.ResourceData{ResourceData: s}
	repo := MessyRemoteRepo{}

	repo.Key = d.GetString("key", false)
	repo.Rclass = "remote"

	repo.RemoteRepoChecksumPolicyType = d.GetString("remote_repo_checksum_policy_type", true)
	repo.AllowAnyHostAuth = d.GetBoolRef("allow_any_host_auth", true)
	repo.BlackedOut = d.GetBoolRef("blacked_out", true)
	repo.BlockMismatchingMimeTypes = d.GetBoolRef("block_mismatching_mime_types", true)
	repo.BowerRegistryUrl = d.GetString("bower_registry_url", true)
	repo.BypassHeadRequests = d.GetBoolRef("bypass_head_requests", true)
	repo.ClientTlsCertificate = d.GetString("client_tls_certificate", true)
	repo.Description = d.GetString("description", true)
	repo.EnableCookieManagement = d.GetBoolRef("enable_cookie_management", true)
	repo.EnableTokenAuthentication = d.GetBool("enable_token_authentication", true)
	repo.ExcludesPattern = d.GetString("excludes_pattern", true)
	repo.FetchJarsEagerly = d.GetBoolRef("fetch_jars_eagerly", true)
	repo.FetchSourcesEagerly = d.GetBoolRef("fetch_sources_eagerly", true)
	repo.HandleReleases = d.GetBoolRef("handle_releases", true)
	repo.HandleSnapshots = d.GetBoolRef("handle_snapshots", true)
	repo.HardFail = d.GetBoolRef("hard_fail", true)
	repo.IncludesPattern = d.GetString("includes_pattern", true)
	repo.LocalAddress = d.GetString("local_address", true)
	repo.MaxUniqueSnapshots = d.GetInt("max_unique_snapshots", true)
	repo.MissedRetrievalCachePeriodSecs = d.GetInt("missed_cache_period_seconds", true)
	repo.Notes = d.GetString("notes", true)
	repo.Offline = d.GetBoolRef("offline", true)
	repo.PackageType = d.GetString("package_type", true)
	repo.Password = d.GetString("password", true)
	repo.PropertySets = d.GetSet("property_sets")
	repo.Proxy = d.GetString("proxy", true)
	repo.PypiRegistryUrl = d.GetString("pypi_registry_url", true)
	repo.RepoLayoutRef = d.GetString("repo_layout_ref", true)
	repo.RetrievalCachePeriodSecs = d.GetInt("retrieval_cache_period_seconds", true)
	repo.ShareConfiguration = d.GetBoolRef("share_configuration", true)
	repo.SocketTimeoutMillis = d.GetInt("socket_timeout_millis", true)
	repo.StoreArtifactsLocally = d.GetBoolRef("store_artifacts_locally", true)
	repo.SuppressPomConsistencyChecks = d.GetBoolRef("suppress_pom_consistency_checks", true)
	repo.SynchronizeProperties = d.GetBoolRef("synchronize_properties", true)
	repo.UnusedArtifactsCleanupPeriodHours = d.GetInt("unused_artifacts_cleanup_period_hours", true)
	repo.Url = d.GetString("url", false)
	repo.Username = d.GetString("username", true)
	repo.VcsGitDownloadUrl = d.GetString("vcs_git_download_url", true)
	repo.VcsGitProvider = d.GetString("vcs_git_provider", true)
	repo.VcsType = d.GetString("vcs_type", true)
	repo.XrayIndex = d.GetBoolRef("xray_index", true)
	repo.FeedContextPath = d.GetString("feed_context_path", true)
	repo.DownloadContextPath = d.GetString("download_context_path", true)
	repo.V3FeedUrl = d.GetString("v3_feed_url", true)
	repo.ForceNugetAuthentication = d.GetBoolRef("force_nuget_authentication", false)
	repo.PropagateQueryParams = d.GetBool("propagate_query_params", true)
	if v, ok := d.GetOk("content_synchronisation"); ok {
		contentSynchronisationConfig := v.([]interface{})[0].(map[string]interface{})
		enabled := contentSynchronisationConfig["enabled"].(bool)
		repo.ContentSynchronisation = &ContentSynchronisation{
			Enabled: enabled,
		}
	}
	if repo.PackageType != "" && repo.PackageType != "generic" && repo.PropagateQueryParams == true {
		format := "cannot use propagate_query_params with repository type %s. This parameter can be used only with generic repositories"
		return MessyRemoteRepo{}, "", fmt.Errorf(format, repo.PackageType)
	}

	return repo, repo.Id(), nil
}

func packLegacyRemoteRepo(r interface{}, d *schema.ResourceData) error {
	repo := r.(*MessyRemoteRepo)
	setValue := util.MkLens(d)

	setValue("remote_repo_checksum_policy_type", repo.RemoteRepoChecksumPolicyType)
	setValue("allow_any_host_auth", repo.AllowAnyHostAuth)
	setValue("blacked_out", repo.BlackedOut)
	setValue("block_mismatching_mime_types", repo.BlockMismatchingMimeTypes)
	setValue("bower_registry_url", repo.BowerRegistryUrl)
	setValue("bypass_head_requests", repo.BypassHeadRequests)
	setValue("client_tls_certificate", repo.ClientTlsCertificate)
	setValue("description", repo.Description)
	setValue("enable_cookie_management", repo.EnableCookieManagement)
	setValue("enable_token_authentication", repo.EnableTokenAuthentication)
	setValue("excludes_pattern", repo.ExcludesPattern)
	setValue("fetch_jars_eagerly", repo.FetchJarsEagerly)
	setValue("fetch_sources_eagerly", repo.FetchSourcesEagerly)
	setValue("handle_releases", repo.HandleReleases)
	setValue("handle_snapshots", repo.HandleSnapshots)
	setValue("hard_fail", repo.HardFail)
	setValue("includes_pattern", repo.IncludesPattern)
	setValue("key", repo.Key)
	setValue("local_address", repo.LocalAddress)
	setValue("max_unique_snapshots", repo.MaxUniqueSnapshots)
	setValue("missed_cache_period_seconds", repo.MissedRetrievalCachePeriodSecs)
	setValue("notes", repo.Notes)
	setValue("offline", repo.Offline)
	setValue("package_type", repo.PackageType)
	setValue("property_sets", schema.NewSet(schema.HashString, util.CastToInterfaceArr(repo.PropertySets)))
	setValue("proxy", repo.Proxy)
	setValue("pypi_registry_url", repo.PypiRegistryUrl)
	setValue("repo_layout_ref", repo.RepoLayoutRef)
	setValue("retrieval_cache_period_seconds", repo.RetrievalCachePeriodSecs)
	setValue("share_configuration", repo.ShareConfiguration)
	setValue("socket_timeout_millis", repo.SocketTimeoutMillis)
	setValue("store_artifacts_locally", repo.StoreArtifactsLocally)
	setValue("suppress_pom_consistency_checks", repo.SuppressPomConsistencyChecks)
	setValue("synchronize_properties", repo.SynchronizeProperties)
	setValue("unused_artifacts_cleanup_period_hours", repo.UnusedArtifactsCleanupPeriodHours)
	setValue("url", repo.Url)
	setValue("username", repo.Username)
	setValue("vcs_git_download_url", repo.VcsGitDownloadUrl)
	setValue("vcs_git_provider", repo.VcsGitProvider)
	setValue("vcs_type", repo.VcsType)
	setValue("xray_index", repo.XrayIndex)
	setValue("feed_context_path", repo.FeedContextPath)
	setValue("download_context_path", repo.DownloadContextPath)
	setValue("v3_feed_url", repo.V3FeedUrl)
	setValue("force_nuget_authentication", repo.ForceNugetAuthentication)
	errors := setValue("propagate_query_params", repo.PropagateQueryParams)
	if repo.ContentSynchronisation != nil {
		setValue("content_synchronisation", []interface{}{
			map[string]bool{
				"enabled": repo.ContentSynchronisation.Enabled,
			},
		})
	}

	if repo.Password != "" {
		errors = setValue("password", util.GetMD5Hash(repo.Password))
	}

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack remote repo %q", errors)
	}
	return nil
}
