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
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				StateFunc: GetMD5Hash,
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
	repo.HardFail = d.GetBoolRef("hard_fail")
	repo.Offline = d.GetBoolRef("offline")
	repo.BlackedOut = d.GetBoolRef("blacked_out")
	repo.StoreArtifactsLocally = d.GetBoolRef("store_artifacts_locally")
	repo.SocketTimeoutMillis = d.GetIntRef("socket_timeout_millis")
	repo.LocalAddress = d.GetStringRef("local_address")
	repo.RetrievalCachePeriodSecs = d.GetIntRef("retrieval_cache_period_seconds")
	repo.MissedRetrievalCachePeriodSecs = d.GetIntRef("missed_cache_period_seconds")
	repo.UnusedArtifactsCleanupPeriodHours = d.GetIntRef("unused_artifacts_cleanup_period_hours")
	repo.ShareConfiguration = d.GetBoolRef("share_configuration")
	repo.SynchronizeProperties = d.GetBoolRef("synchronize_properties")
	repo.BlockMismatchingMimeTypes = d.GetBoolRef("block_mismatching_mime_types")
	repo.AllowAnyHostAuth = d.GetBoolRef("allow_any_host_auth")
	repo.EnableCookieManagement = d.GetBoolRef("enable_cookie_management")
	repo.ClientTLSCertificate = d.GetStringRef("client_tls_certificate")
	repo.PropertySets = d.GetSetRef("property_sets")
	repo.HandleReleases = d.GetBoolRef("handle_releases")
	repo.HandleSnapshots = d.GetBoolRef("handle_snapshots")
	//repo.RemoteRepoChecksumPolicyType = d.GetStringRef("remote_repo_checksum_policy_type")
	repo.MaxUniqueSnapshots = d.GetIntRef("max_unique_snapshots")
	repo.SuppressPomConsistencyChecks = d.GetBoolRef("suppress_pom_consistency_checks")
	repo.FetchJarsEagerly = d.GetBoolRef("fetch_jars_eagerly")
	repo.FetchSourcesEagerly = d.GetBoolRef("fetch_sources_eagerly")
	repo.PyPiRegistryUrl = d.GetStringRef("pypi_registry_url")
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
	d.Set("password", GetMD5Hash(*repo.Password))
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
	d.Set("property_sets", schema.NewSet(schema.HashString, CastToInterfaceArr(*repo.PropertySets)))
	d.Set("handle_releases", repo.HandleReleases)
	d.Set("handle_snapshots", repo.HandleSnapshots)
	//d.Set("remote_repo_checksum_policy_type", repo.RemoteRepoChecksumPolicyType)
	d.Set("max_unique_snapshots", repo.MaxUniqueSnapshots)
	d.Set("fetch_jars_eagerly", repo.FetchJarsEagerly)
	d.Set("fetch_sources_eagerly", repo.FetchSourcesEagerly)
	d.Set("pypi_registry_url", repo.PyPiRegistryUrl)
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
