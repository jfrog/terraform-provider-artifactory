package artifactory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"package_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"suppress_pom_consistency_checks": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"local_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"retrieval_cache_period_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"missed_cache_period_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"unused_artifacts_cleanup_period_hours": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
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
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"nuget"},
			},
			"download_context_path": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"nuget"},
			},
			"v3_feed_url": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"nuget"},
			},
			"content_synchronisation": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"nuget": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				MinItems:      1,
				Deprecated:    "Since Artifactory 6.9.0+ (provider 1.6). Use /api/v2 endpoint",
				ConflictsWith: []string{"feed_context_path", "download_context_path", "v3_feed_url"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
					},
				},
			},
		},
	}
}

func unpackRemoteRepo(s *schema.ResourceData) *v1.RemoteRepository {
	d := &ResourceData{s}
	repo := new(v1.RemoteRepository)

	repo.Key = d.getStringRef("key", false)
	repo.RClass = artifactory.String("remote")

	repo.RemoteRepoChecksumPolicyType = d.getStringRef("remote_repo_checksum_policy_type", true)
	repo.AllowAnyHostAuth = d.getBoolRef("allow_any_host_auth", true)
	repo.BlackedOut = d.getBoolRef("blacked_out", true)
	repo.BlockMismatchingMimeTypes = d.getBoolRef("block_mismatching_mime_types", true)
	repo.BowerRegistryURL = d.getStringRef("bower_registry_url", true)
	repo.BypassHeadRequests = d.getBoolRef("bypass_head_requests", true)
	repo.ClientTLSCertificate = d.getStringRef("client_tls_certificate", true)
	repo.Description = d.getStringRef("description", true)
	repo.EnableCookieManagement = d.getBoolRef("enable_cookie_management", true)
	repo.EnableTokenAuthentication = d.getBoolRef("enable_token_authentication", true)
	repo.ExcludesPattern = d.getStringRef("excludes_pattern", true)
	repo.FetchJarsEagerly = d.getBoolRef("fetch_jars_eagerly", true)
	repo.FetchSourcesEagerly = d.getBoolRef("fetch_sources_eagerly", true)
	repo.HandleReleases = d.getBoolRef("handle_releases", true)
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots", true)
	repo.HardFail = d.getBoolRef("hard_fail", true)
	repo.IncludesPattern = d.getStringRef("includes_pattern", true)
	repo.LocalAddress = d.getStringRef("local_address", true)
	repo.MaxUniqueSnapshots = d.getIntRef("max_unique_snapshots", true)
	repo.MissedRetrievalCachePeriodSecs = d.getIntRef("missed_cache_period_seconds", true)
	repo.Notes = d.getStringRef("notes", true)
	repo.Offline = d.getBoolRef("offline", true)
	repo.PackageType = d.getStringRef("package_type", true)
	repo.Password = d.getStringRef("password", true)
	repo.PropertySets = d.getSetRef("property_sets")
	repo.Proxy = d.getStringRef("proxy", true)
	repo.PyPiRegistryUrl = d.getStringRef("pypi_registry_url", true)
	repo.RepoLayoutRef = d.getStringRef("repo_layout_ref", true)
	repo.RetrievalCachePeriodSecs = d.getIntRef("retrieval_cache_period_seconds", true)
	repo.ShareConfiguration = d.getBoolRef("share_configuration", true)
	repo.SocketTimeoutMillis = d.getIntRef("socket_timeout_millis", true)
	repo.StoreArtifactsLocally = d.getBoolRef("store_artifacts_locally", true)
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks", true)
	repo.SynchronizeProperties = d.getBoolRef("synchronize_properties", true)
	repo.UnusedArtifactsCleanupPeriodHours = d.getIntRef("unused_artifacts_cleanup_period_hours", true)
	repo.Url = d.getStringRef("url", true)
	repo.Username = d.getStringRef("username", true)
	repo.VcsGitDownloadUrl = d.getStringRef("vcs_git_download_url", true)
	repo.VcsGitProvider = d.getStringRef("vcs_git_provider", true)
	repo.VcsType = d.getStringRef("vcs_type", true)
	repo.XrayIndex = d.getBoolRef("xray_index", true)
	repo.FeedContextPath = d.getStringRef("feed_context_path", true)
	repo.DownloadContextPath = d.getStringRef("download_context_path", true)
	repo.V3FeedUrl = d.getStringRef("v3_feed_url", true)
	if v, ok := d.GetOk("content_synchronisation"); ok {
		contentSynchronisationConfig := v.([]interface{})[0].(map[string]interface{})
		enabled := contentSynchronisationConfig["enabled"].(bool)
		repo.ContentSynchronisation = &v1.ContentSynchronisation{
			Enabled: &enabled,
		}
	}
	if v, ok := d.GetOk("nuget"); ok {
		nugetConfig := v.([]interface{})[0].(map[string]interface{})
		feedContextPath := nugetConfig["feed_context_path"].(string)
		downloadContextPath := nugetConfig["download_context_path"].(string)
		v3FeedUrl := nugetConfig["v3_feed_url"].(string)
		repo.Nuget = &v1.Nuget{
			FeedContextPath:     &feedContextPath,
			DownloadContextPath: &downloadContextPath,
			V3FeedUrl:           &v3FeedUrl,
		}
	}

	return repo
}

func packRemoteRepo(repo *v1.RemoteRepository, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("remote_repo_checksum_policy_type", repo.RemoteRepoChecksumPolicyType))
	logErr(d.Set("allow_any_host_auth", repo.AllowAnyHostAuth))
	logErr(d.Set("blacked_out", repo.BlackedOut))
	logErr(d.Set("block_mismatching_mime_types", repo.BlockMismatchingMimeTypes))
	logErr(d.Set("bower_registry_url", repo.BowerRegistryURL))
	logErr(d.Set("bypass_head_requests", repo.BypassHeadRequests))
	logErr(d.Set("client_tls_certificate", repo.ClientTLSCertificate))
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
	logErr(d.Set("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(*repo.PropertySets))))
	logErr(d.Set("proxy", repo.Proxy))
	logErr(d.Set("pypi_registry_url", repo.PyPiRegistryUrl))
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
	if repo.ContentSynchronisation != nil {
		logErr(d.Set("content_synchronisation", []interface{}{
			map[string]*bool{
				"enabled": repo.ContentSynchronisation.Enabled,
			},
		}))
	}
	if repo.Nuget != nil {
		logErr(d.Set("nuget", []interface{}{
			map[string]*string{
				"feed_context_path":     repo.Nuget.FeedContextPath,
				"download_context_path": repo.Nuget.DownloadContextPath,
				"v3_feed_url":           repo.Nuget.V3FeedUrl,
			},
		}))
	}

	if repo.Password != nil {
		logErr(d.Set("password", getMD5Hash(*repo.Password)))
	}

	if hasErr {
		return fmt.Errorf("failed to pack remote repo")
	}
	return nil
}

func resourceRemoteRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	repo := unpackRemoteRepo(d)
	_, err := c.V1.Repositories.CreateRemote(context.Background(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	repo, resp, err := c.V1.Repositories.GetRemote(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	return packRemoteRepo(repo, d)
}

func resourceRemoteRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	repo := unpackRemoteRepo(d)

	_, err := c.V1.Repositories.UpdateRemote(context.Background(), d.Id(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)
	repo := unpackRemoteRepo(d)

	resp, err := c.V1.Repositories.DeleteRemote(context.Background(), *repo.Key)
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	return err
}

func resourceRemoteRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*artifactory.Artifactory)

	key := d.Id()
	_, resp, err := c.V1.Repositories.GetRemote(context.Background(), key)

	// Cannot check for 404 because artifactory returns 400
	if resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}

	return true, err
}
