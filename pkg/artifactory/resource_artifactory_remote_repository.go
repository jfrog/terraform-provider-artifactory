package artifactory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
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
				Default:  "generic",
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
				Default:  "**/*",
			},
			"excludes_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"repo_layout_ref": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"handle_releases": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"handle_snapshots": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"max_unique_snapshots": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"suppress_pom_consistency_checks": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
				Default:  "generate-if-absent",
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
				Default:  false,
			},
			"offline": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"blacked_out": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"store_artifacts_locally": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"socket_timeout_millis": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  15000,
			},
			"local_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"retrieval_cache_period_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  43200,
			},
			"missed_cache_period_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  7200,
			},
			"unused_artifacts_cleanup_period_hours": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"fetch_jars_eagerly": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"fetch_sources_eagerly": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"share_configuration": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"synchronize_properties": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"block_mismatching_mime_types": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
				Default:  false,
			},
			"enable_cookie_management": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"client_tls_certificate": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"pypi_registry_url": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"bower_registry_url": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"bypass_head_requests": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_token_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"xray_index": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"vcs_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"vcs_git_provider": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"vcs_git_download_url": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"nuget": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"feed_context_path": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "api/v2",
						},
						"download_context_path": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "api/v2/package",
						},
						"v3_feed_url": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "https://api.nuget.org/v3/index.json",
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

	repo.RemoteRepoChecksumPolicyType = d.getStringRef("remote_repo_checksum_policy_type")
	repo.AllowAnyHostAuth = d.getBoolRef("allow_any_host_auth")
	repo.BlackedOut = d.getBoolRef("blacked_out")
	repo.BlockMismatchingMimeTypes = d.getBoolRef("block_mismatching_mime_types")
	repo.BowerRegistryURL = d.getStringRef("bower_registry_url")
	repo.BypassHeadRequests = d.getBoolRef("bypass_head_requests")
	repo.ClientTLSCertificate = d.getStringRef("client_tls_certificate")
	repo.Description = d.getStringRef("description")
	repo.EnableCookieManagement = d.getBoolRef("enable_cookie_management")
	repo.EnableTokenAuthentication = d.getBoolRef("enable_token_authentication")
	repo.ExcludesPattern = d.getStringRef("excludes_pattern")
	repo.FetchJarsEagerly = d.getBoolRef("fetch_jars_eagerly")
	repo.FetchSourcesEagerly = d.getBoolRef("fetch_sources_eagerly")
	repo.HandleReleases = d.getBoolRef("handle_releases")
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots")
	repo.HardFail = d.getBoolRef("hard_fail")
	repo.IncludesPattern = d.getStringRef("includes_pattern")
	repo.Key = d.getStringRef("key")
	repo.LocalAddress = d.getStringRef("local_address")
	repo.MaxUniqueSnapshots = d.getIntRef("max_unique_snapshots")
	repo.MissedRetrievalCachePeriodSecs = d.getIntRef("missed_cache_period_seconds")
	repo.Notes = d.getStringRef("notes")
	repo.Offline = d.getBoolRef("offline")
	repo.PackageType = d.getStringRef("package_type")
	repo.Password = d.getStringRef("password")
	repo.PropertySets = d.getSetRef("property_sets")
	repo.Proxy = d.getStringRef("proxy")
	repo.PyPiRegistryUrl = d.getStringRef("pypi_registry_url")
	repo.RClass = artifactory.String("remote")
	repo.RepoLayoutRef = d.getStringRef("repo_layout_ref")
	repo.RetrievalCachePeriodSecs = d.getIntRef("retrieval_cache_period_seconds")
	repo.ShareConfiguration = d.getBoolRef("share_configuration")
	repo.SocketTimeoutMillis = d.getIntRef("socket_timeout_millis")
	repo.StoreArtifactsLocally = d.getBoolRef("store_artifacts_locally")
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks")
	repo.SynchronizeProperties = d.getBoolRef("synchronize_properties")
	repo.UnusedArtifactsCleanupPeriodHours = d.getIntRef("unused_artifacts_cleanup_period_hours")
	repo.Url = d.getStringRef("url")
	repo.Username = d.getStringRef("username")
	repo.VcsGitDownloadUrl = d.getStringRef("vcs_git_download_url")
	repo.VcsGitProvider = d.getStringRef("vcs_git_provider")
	repo.VcsType = d.getStringRef("vcs_type")
	repo.XrayIndex = d.getBoolRef("xray_index")
	if v, ok := d.GetOk("nuget"); *repo.PackageType == "nuget" && ok {
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
