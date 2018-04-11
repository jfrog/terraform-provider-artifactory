package artifactory

import (
	"fmt"

	"context"
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
				Computed: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == fmt.Sprintf("%s (local file cache)", new)
				},
			},
			"notes": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"includes_pattern": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"excludes_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"repo_layout_ref": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"handle_releases": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
			},
			"handle_snapshots": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
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
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				Computed:  true,
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
				Computed: true,
				Optional: true,
			},
			"socket_timeout_millis": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"local_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"retrieval_cache_period_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"failed_cache_period_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"missed_cache_period_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"unused_artifacts_cleanup_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"unused_artifacts_cleanup_period_hours": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
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
				Computed: true,
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
		},
	}
}

func unmarshalRemoteRepository(s *schema.ResourceData) *artifactory.RemoteRepository {
	d := &ResourceData{s}
	repo := new(artifactory.RemoteRepository)

	repo.Key = d.GetStringRef("key")
	repo.RClass = artifactory.String("remote")
	repo.PackageType = d.GetStringRef("package_type")
	repo.Url = d.GetStringRef("url")
	repo.Proxy = d.GetStringRef("proxy")
	repo.Username = d.GetStringRef("username")
	repo.Password = d.GetStringRef("password")
	repo.Description = d.GetStringRef("description")
	repo.Notes = d.GetStringRef("notes")
	repo.IncludesPattern = d.GetStringRef("includes_pattern")
	repo.ExcludesPattern = d.GetStringRef("excludes_pattern")
	repo.RepoLayoutRef = d.GetStringRef("repo_layout_ref")
	repo.RemoteRepoChecksumPolicyType = d.GetStringRef("remote_repo_checksum_policy_type")
	repo.HandleReleases = d.GetBoolRef("handle_releases")
	repo.HandleSnapshots = d.GetBoolRef("handle_snapshots")
	repo.MaxUniqueSnapshots = d.GetIntRef("max_unique_snapshots")
	repo.SuppressPomConsistencyChecks = d.GetBoolRef("suppress_pom_consistency_checks")
	repo.HardFail = d.GetBoolRef("hard_fail")
	repo.Offline = d.GetBoolRef("offline")
	repo.BlackedOut = d.GetBoolRef("blacked_out")
	repo.StoreArtifactsLocally = d.GetBoolRef("store_artifacts_locally")
	repo.SocketTimeoutMillis = d.GetIntRef("socket_timeout_millis")
	repo.LocalAddress = d.GetStringRef("local_address")
	repo.RetrievalCachePeriodSecs = d.GetIntRef("retrieval_cache_period_seconds")
	repo.FailedRetrievalCachePeriodSecs = d.GetIntRef("failed_cache_period_seconds")
	repo.MissedRetrievalCachePeriodSecs = d.GetIntRef("missed_cache_period_seconds")
	repo.UnusedArtifactsCleanupEnabled = d.GetBoolRef("unused_artifacts_cleanup_enabled")
	repo.UnusedArtifactsCleanupPeriodHours = d.GetIntRef("unused_artifacts_cleanup_period_hours")
	repo.FetchJarsEagerly = d.GetBoolRef("fetch_jars_eagerly")
	repo.FetchSourcesEagerly = d.GetBoolRef("fetch_sources_eagerly")
	repo.ShareConfiguration = d.GetBoolRef("share_configuration")
	repo.SynchronizeProperties = d.GetBoolRef("synchronize_properties")
	repo.BlockMismatchingMimeTypes = d.GetBoolRef("block_mismatching_mime_types")
	repo.AllowAnyHostAuth = d.GetBoolRef("allow_any_host_auth")
	repo.EnableCookieManagement = d.GetBoolRef("enable_cookie_management")
	repo.ClientTLSCertificate = d.GetStringRef("client_tls_certificate")
	repo.PropertySets = d.GetSetRef("property_sets")

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
	d.Set("handle_releases", repo.HandleReleases)
	d.Set("handle_snapshots", repo.HandleSnapshots)
	d.Set("max_unique_snapshots", repo.MaxUniqueSnapshots)
	d.Set("suppress_pom_consistency_checks", repo.SuppressPomConsistencyChecks)
	d.Set("blacked_out", repo.BlackedOut)
	d.Set("url", repo.Url)
	d.Set("username", repo.Username)
	d.Set("proxy", repo.Proxy)
	d.Set("remote_repo_checksum_policy_type", repo.RemoteRepoChecksumPolicyType)
	d.Set("hard_fail", repo.HardFail)
	d.Set("offline", repo.Offline)
	d.Set("store_artifacts_locally", repo.StoreArtifactsLocally)
	d.Set("socket_timeout_millis", repo.SocketTimeoutMillis)
	d.Set("local_address", repo.LocalAddress)
	d.Set("retrieval_cache_period_seconds", repo.RetrievalCachePeriodSecs)
	d.Set("failed_cache_period_seconds", repo.FailedRetrievalCachePeriodSecs)
	d.Set("missed_cache_period_seconds", repo.MissedRetrievalCachePeriodSecs)
	d.Set("unused_artifacts_cleanup_enabled", repo.UnusedArtifactsCleanupEnabled)
	d.Set("unused_artifacts_cleanup_period_hours", repo.UnusedArtifactsCleanupPeriodHours)
	d.Set("fetch_jars_eagerly", repo.FetchJarsEagerly)
	d.Set("fetch_sources_eagerly", repo.FetchSourcesEagerly)
	d.Set("share_configuration", repo.ShareConfiguration)
	d.Set("synchronize_properties", repo.SynchronizeProperties)
	d.Set("block_mismatching_mime_types", repo.BlockMismatchingMimeTypes)
	d.Set("allow_any_host_auth", repo.AllowAnyHostAuth)
	d.Set("enable_cookie_management", repo.EnableCookieManagement)
	d.Set("client_tls_certificate", repo.ClientTLSCertificate)
	d.Set("property_sets", repo.PropertySets)
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
