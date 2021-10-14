package artifactory

import (
	"fmt"

	"github.com/go-resty/resty/v2"

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
	repo.Url = d.getString("url", false)
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
	setValue := mkLens(d)

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
	setValue("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(repo.PropertySets)))
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
		errors = setValue("password", getMD5Hash(repo.Password))
	}

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack remote repo %q", errors)
	}
	return nil
}

func resourceRemoteRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	repo, err := unpackRemoteRepo(d)
	if err != nil {
		return err
	}

	_, err = m.(*resty.Client).R().SetBody(repo).Put(repositoriesEndpoint + repo.Key)
	if err != nil {
		return err
	}

	d.SetId(repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryRead(d *schema.ResourceData, m interface{}) error {
	repo := MessyRemoteRepo{}
	resp, err := m.(*resty.Client).R().SetResult(&repo).Get(repositoriesEndpoint + d.Id())
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
	repo, err := unpackRemoteRepo(d)
	if err != nil {
		return err
	}
	_, err = m.(*resty.Client).R().SetBody(repo).Post(repositoriesEndpoint + repo.Key)
	if err != nil {
		return err
	}

	d.SetId(repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	resp, err := m.(*resty.Client).R().Delete(repositoriesEndpoint + d.Id())

	if err != nil {
		if resp != nil && resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}

	return err
}

func resourceRemoteRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := m.(*resty.Client).R().Head(repositoriesEndpoint + d.Id())

	// as long as we don't have an error, it's good
	return err == nil, err
}
