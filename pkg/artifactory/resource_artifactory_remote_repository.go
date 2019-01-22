package artifactory

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
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
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
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
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				StateFunc:        getMD5Hash,
				DiffSuppressFunc: mD5Diff,
			},
			"proxy": {
				Type:     schema.TypeString,
				Optional: true,
			},
			/*"remote_repo_checksum_policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "generate_if_absent",
				Removed:  "since sometime",
			},*/
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
		},
	}
}

func unmarshalRemoteRepository(s *schema.ResourceData) *artifactory.RemoteRepository {
	d := &ResourceData{s}
	repo := new(artifactory.RemoteRepository)

	repo.Key = d.getStringRef("key")
	repo.RClass = artifactory.String("remote")
	repo.PackageType = d.getStringRef("package_type")
	repo.Url = d.getStringRef("url")
	repo.Proxy = d.getStringRef("proxy")
	repo.Username = d.getStringRef("username")
	repo.Password = d.getStringRef("password")
	repo.Description = d.getStringRef("description")
	repo.Notes = d.getStringRef("notes")
	repo.IncludesPattern = d.getStringRef("includes_pattern")
	repo.ExcludesPattern = d.getStringRef("excludes_pattern")
	repo.RepoLayoutRef = d.getStringRef("repo_layout_ref")
	repo.HardFail = d.getBoolRef("hard_fail")
	repo.Offline = d.getBoolRef("offline")
	repo.BlackedOut = d.getBoolRef("blacked_out")
	repo.StoreArtifactsLocally = d.getBoolRef("store_artifacts_locally")
	repo.SocketTimeoutMillis = d.getIntRef("socket_timeout_millis")
	repo.LocalAddress = d.getStringRef("local_address")
	repo.RetrievalCachePeriodSecs = d.getIntRef("retrieval_cache_period_seconds")
	repo.MissedRetrievalCachePeriodSecs = d.getIntRef("missed_cache_period_seconds")
	repo.UnusedArtifactsCleanupPeriodHours = d.getIntRef("unused_artifacts_cleanup_period_hours")
	repo.ShareConfiguration = d.getBoolRef("share_configuration")
	repo.SynchronizeProperties = d.getBoolRef("synchronize_properties")
	repo.BlockMismatchingMimeTypes = d.getBoolRef("block_mismatching_mime_types")
	repo.AllowAnyHostAuth = d.getBoolRef("allow_any_host_auth")
	repo.EnableCookieManagement = d.getBoolRef("enable_cookie_management")
	repo.ClientTLSCertificate = d.getStringRef("client_tls_certificate")
	repo.PropertySets = d.getSetRef("property_sets")
	repo.HandleReleases = d.getBoolRef("handle_releases")
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots")
	//repo.RemoteRepoChecksumPolicyType = d.getStringRef("remote_repo_checksum_policy_type")
	repo.MaxUniqueSnapshots = d.getIntRef("max_unique_snapshots")
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks")
	repo.FetchJarsEagerly = d.getBoolRef("fetch_jars_eagerly")
	repo.FetchSourcesEagerly = d.getBoolRef("fetch_sources_eagerly")
	repo.PyPiRegistryUrl = d.getStringRef("pypi_registry_url")
	repo.BypassHeadRequests = d.getBoolRef("bypass_head_requests")
	repo.EnableTokenAuthentication = d.getBoolRef("enable_token_authentication")
	return repo
}

func marshalRemoteRepository(repo *artifactory.RemoteRepository, d *schema.ResourceData) {
	d.Set("key", repo.Key)
	d.Set("type", repo.RClass)
	d.Set("package_type", repo.PackageType)
	d.Set("description", repo.Description)
	d.Set("notes", repo.Notes)
	d.Set("includes_pattern", repo.IncludesPattern)
	d.Set("excludes_pattern", repo.ExcludesPattern)
	d.Set("repo_layout_ref", repo.RepoLayoutRef)
	d.Set("blacked_out", repo.BlackedOut)
	d.Set("url", repo.Url)
	d.Set("username", repo.Username)
	d.Set("password", *repo.Password)
	d.Set("proxy", repo.Proxy)
	d.Set("hard_fail", repo.HardFail)
	d.Set("offline", repo.Offline)
	d.Set("store_artifacts_locally", repo.StoreArtifactsLocally)
	d.Set("socket_timeout_millis", repo.SocketTimeoutMillis)
	d.Set("local_address", repo.LocalAddress)
	d.Set("retrieval_cache_period_seconds", repo.RetrievalCachePeriodSecs)
	d.Set("missed_cache_period_seconds", repo.MissedRetrievalCachePeriodSecs)
	d.Set("unused_artifacts_cleanup_period_hours", repo.UnusedArtifactsCleanupPeriodHours)
	d.Set("share_configuration", repo.ShareConfiguration)
	d.Set("synchronize_properties", repo.SynchronizeProperties)
	d.Set("block_mismatching_mime_types", repo.BlockMismatchingMimeTypes)
	d.Set("allow_any_host_auth", repo.AllowAnyHostAuth)
	d.Set("enable_cookie_management", repo.EnableCookieManagement)
	d.Set("client_tls_certificate", repo.ClientTLSCertificate)
	d.Set("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(*repo.PropertySets)))
	d.Set("handle_releases", repo.HandleReleases)
	d.Set("handle_snapshots", repo.HandleSnapshots)
	//d.Set("remote_repo_checksum_policy_type", repo.RemoteRepoChecksumPolicyType)
	d.Set("max_unique_snapshots", repo.MaxUniqueSnapshots)
	d.Set("fetch_jars_eagerly", repo.FetchJarsEagerly)
	d.Set("fetch_sources_eagerly", repo.FetchSourcesEagerly)
	d.Set("pypi_registry_url", repo.PyPiRegistryUrl)
	d.Set("bypass_head_requests", repo.BypassHeadRequests)
	d.Set("enable_token_authentication", repo.EnableTokenAuthentication)
}

func resourceRemoteRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	repo := unmarshalRemoteRepository(d)
	_, err := c.Repositories.CreateRemote(context.Background(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	repo, resp, err := c.Repositories.GetRemote(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	marshalRemoteRepository(repo, d)
	return nil
}

func resourceRemoteRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	repo := unmarshalRemoteRepository(d)

	_, err := c.Repositories.UpdateRemote(context.Background(), d.Id(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceRemoteRepositoryRead(d, m)
}

func resourceRemoteRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)
	repo := unmarshalRemoteRepository(d)

	resp, err := c.Repositories.DeleteRemote(context.Background(), *repo.Key)
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	return err
}

func resourceRemoteRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*artifactory.Client)

	key := d.Id()
	_, resp, err := c.Repositories.GetRemote(context.Background(), key)

	// Cannot check for 404 because artifactory returns 400
	if resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}

	return true, err
}
