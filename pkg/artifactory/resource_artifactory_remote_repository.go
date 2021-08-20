package artifactory

import (
	"fmt"
	"net/http"

	"github.com/jfrog/jfrog-client-go/artifactory/services"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type MessyRemoteRepo struct {
	services.RemoteRepositoryBaseParams
	services.BowerRemoteRepositoryParams
	services.CommonMavenGradleRemoteRepositoryParams
	services.DockerRemoteRepositoryParams
	services.VcsRemoteRepositoryParams
	services.PypiRemoteRepositoryParams
	services.NugetRemoteRepositoryParams
	PropagateQueryParams bool `json:"propagateQueryParams"`
}

func resourceArtifactoryRemoteRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceRemoteRepositoryCreate,
		Read:   resourceRemoteRepositoryRead,
		Update: resourceRemoteRepositoryUpdate,
		Delete: resourceRemoteRepositoryDelete,
		Exists: resourceRemoteRepositoryExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: repoKeyValidator,
			},
			"package_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "generic",
				ValidateFunc: repoTypeValidator,
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
		},
	}
}

func unpackRemoteRepo(s *schema.ResourceData) (MessyRemoteRepo, error) {
	d := &ResourceData{s}
	repo := MessyRemoteRepo{}

	repo.Key = d.getString("key", false)
	repo.Rclass = "remote"

	repo.RemoteRepoChecksumPolicyType = d.getString("remote_repo_checksum_policy_type", true)
	repo.AllowAnyHostAuth = d.getBoolRef("allow_any_host_auth", true)
	repo.BlackedOut = d.getBoolRef("blacked_out", true)
	repo.BlockMismatchingMimeTypes = d.getBoolRef("block_mismatching_mime_types", true)
	repo.BowerRegistryUrl = d.getString("bower_registry_url", true)
	repo.BypassHeadRequests = d.getBoolRef("bypass_head_requests", true)
	repo.ClientTlsCertificate = d.getString("client_tls_certificate", true)
	repo.Description = d.getString("description", true)
	repo.EnableCookieManagement = d.getBoolRef("enable_cookie_management", true)
	repo.EnableTokenAuthentication = d.getBoolRef("enable_token_authentication", true)
	repo.ExcludesPattern = d.getString("excludes_pattern", true)
	repo.FetchJarsEagerly = d.getBoolRef("fetch_jars_eagerly", true)
	repo.FetchSourcesEagerly = d.getBoolRef("fetch_sources_eagerly", true)
	repo.HandleReleases = d.getBoolRef("handle_releases", true)
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots", true)
	repo.HardFail = d.getBoolRef("hard_fail", true)
	repo.IncludesPattern = d.getString("includes_pattern", true)
	repo.LocalAddress = d.getString("local_address", true)
	repo.MaxUniqueSnapshots = d.getInt("max_unique_snapshots", true)
	repo.MissedRetrievalCachePeriodSecs = d.getInt("missed_cache_period_seconds", true)
	repo.Notes = d.getString("notes", true)
	repo.Offline = d.getBoolRef("offline", true)
	repo.PackageType = d.getString("package_type", true)
	repo.Password = d.getString("password", true)
	repo.PropertySets = d.getSet("property_sets")
	repo.Proxy = d.getString("proxy", true)
	repo.PypiRegistryUrl = d.getString("pypi_registry_url", true)
	repo.RepoLayoutRef = d.getString("repo_layout_ref", true)
	repo.RetrievalCachePeriodSecs = d.getInt("retrieval_cache_period_seconds", true)
	repo.ShareConfiguration = d.getBoolRef("share_configuration", true)
	repo.SocketTimeoutMillis = d.getInt("socket_timeout_millis", true)
	repo.StoreArtifactsLocally = d.getBoolRef("store_artifacts_locally", true)
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks", true)
	repo.SynchronizeProperties = d.getBoolRef("synchronize_properties", true)
	repo.UnusedArtifactsCleanupPeriodHours = d.getInt("unused_artifacts_cleanup_period_hours", true)
	repo.Url = d.getString("url", true)
	repo.Username = d.getString("username", true)
	repo.VcsGitDownloadUrl = d.getString("vcs_git_download_url", true)
	repo.VcsGitProvider = d.getString("vcs_git_provider", true)
	repo.VcsType = d.getString("vcs_type", true)
	repo.XrayIndex = d.getBoolRef("xray_index", true)
	repo.FeedContextPath = d.getString("feed_context_path", true)
	repo.DownloadContextPath = d.getString("download_context_path", true)
	repo.V3FeedUrl = d.getString("v3_feed_url", true)
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)
	repo.PropagateQueryParams = d.getBool("propagate_query_params", true)
	if v, ok := d.GetOk("content_synchronisation"); ok {
		contentSynchronisationConfig := v.([]interface{})[0].(map[string]interface{})
		enabled := contentSynchronisationConfig["enabled"].(bool)
		repo.ContentSynchronisation = &services.ContentSynchronisation{
			Enabled: enabled,
		}
	}
	if repo.PackageType != "" && repo.PackageType != "generic" && repo.PropagateQueryParams == true {
		return MessyRemoteRepo{}, fmt.Errorf("cannot use propagate_query_params with repository type %s. This parameter can be used only with generic repositories", repo.PackageType)
	}

	return repo, nil
}

func packRemoteRepo(repo MessyRemoteRepo, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("remote_repo_checksum_policy_type", repo.RemoteRepoChecksumPolicyType))
	logErr(d.Set("allow_any_host_auth", repo.AllowAnyHostAuth))
	logErr(d.Set("blacked_out", repo.BlackedOut))
	logErr(d.Set("block_mismatching_mime_types", repo.BlockMismatchingMimeTypes))
	logErr(d.Set("bower_registry_url", repo.BowerRegistryUrl))
	logErr(d.Set("bypass_head_requests", repo.BypassHeadRequests))
	logErr(d.Set("client_tls_certificate", repo.ClientTlsCertificate))
	logErr(d.Set("description", repo.Description))
	logErr(d.Set("enable_cookie_management", repo.EnableCookieManagement))
	logErr(d.Set("enable_token_authentication", repo.EnableTokenAuthentication))
	logErr(d.Set("excludes_pattern", repo.ExcludesPattern))
	logErr(d.Set("fetch_jars_eagerly", repo.FetchJarsEagerly))
	logErr(d.Set("fetch_sources_eagerly", repo.FetchSourcesEagerly))
	logErr(d.Set("handle_releases", repo.HandleReleases))
	logErr(d.Set("handle_snapshots", repo.HandleSnapshots))
	logErr(d.Set("hard_fail", repo.HardFail))
	logErr(d.Set("includes_pattern", repo.IncludesPattern))
	logErr(d.Set("key", repo.Key))
	logErr(d.Set("local_address", repo.LocalAddress))
	logErr(d.Set("max_unique_snapshots", repo.MaxUniqueSnapshots))
	logErr(d.Set("missed_cache_period_seconds", repo.MissedRetrievalCachePeriodSecs))
	logErr(d.Set("notes", repo.Notes))
	logErr(d.Set("offline", repo.Offline))
	logErr(d.Set("package_type", repo.PackageType))
	logErr(d.Set("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(repo.PropertySets))))
	logErr(d.Set("proxy", repo.Proxy))
	logErr(d.Set("pypi_registry_url", repo.PypiRegistryUrl))
	logErr(d.Set("repo_layout_ref", repo.RepoLayoutRef))
	logErr(d.Set("retrieval_cache_period_seconds", repo.RetrievalCachePeriodSecs))
	logErr(d.Set("share_configuration", repo.ShareConfiguration))
	logErr(d.Set("socket_timeout_millis", repo.SocketTimeoutMillis))
	logErr(d.Set("store_artifacts_locally", repo.StoreArtifactsLocally))
	logErr(d.Set("suppress_pom_consistency_checks", repo.SuppressPomConsistencyChecks))
	logErr(d.Set("synchronize_properties", repo.SynchronizeProperties))
	logErr(d.Set("unused_artifacts_cleanup_period_hours", repo.UnusedArtifactsCleanupPeriodHours))
	logErr(d.Set("url", repo.Url))
	logErr(d.Set("username", repo.Username))
	logErr(d.Set("vcs_git_download_url", repo.VcsGitDownloadUrl))
	logErr(d.Set("vcs_git_provider", repo.VcsGitProvider))
	logErr(d.Set("vcs_type", repo.VcsType))
	logErr(d.Set("xray_index", repo.XrayIndex))
	logErr(d.Set("feed_context_path", repo.FeedContextPath))
	logErr(d.Set("download_context_path", repo.DownloadContextPath))
	logErr(d.Set("v3_feed_url", repo.V3FeedUrl))
	logErr(d.Set("force_nuget_authentication", repo.ForceNugetAuthentication))
	logErr(d.Set("propagate_query_params", repo.PropagateQueryParams))
	if repo.ContentSynchronisation != nil {
		logErr(d.Set("content_synchronisation", []interface{}{
			map[string]bool{
				"enabled": repo.ContentSynchronisation.Enabled,
			},
		}))
	}

	if repo.Password != "" {
		logErr(d.Set("password", getMD5Hash(repo.Password)))
	}

	if hasErr {
		return fmt.Errorf("failed to pack remote repo")
	}
	return nil
}

func resourceRemoteRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	repo, err := unpackRemoteRepo(d)
	if err != nil {
		return err
	}

	_, err = client.R().SetBody(repo).Put("artifactory/api/repositories/" + repo.Key)
	if err != nil {
		return err
	}

	d.SetId(repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty
	repo := MessyRemoteRepo{}
	resp, err := client.R().SetResult(&repo).Get("artifactory/api/repositories/" + d.Id())
	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}
	if resp == nil {
		return fmt.Errorf("no response returned during resourceRemoteRepositoryRead")
	}

	return packRemoteRepo(repo, d)
}

func resourceRemoteRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	repo, err := unpackRemoteRepo(d)
	if err != nil {
		return err
	}
	_, err = client.R().SetBody(repo).Post("artifactory/api/repositories/" + repo.Key)
	if err != nil {
		return err
	}

	d.SetId(repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty
	repo, err := unpackRemoteRepo(d)
	if err != nil {
		return err
	}
	resp, err := client.R().Delete("artifactory/api/repositories/" + repo.Key)

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	return err
}

func resourceRemoteRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ArtClient).Resty

	_, err := client.R().Head("artifactory/api/repositories/" + d.Id())

	// as long as we don't have an error, it's good
	return err == nil, err
}
