package artifactory

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"net/http"
	"regexp"
)

const repositoriesEndpoint = "artifactory/api/repositories/"

type ReadFunc func(d *schema.ResourceData, m interface{}) error

// Constructor Must return a pointer to a struct. When just returning a struct, resty gets confused and thinks it's a map
type Constructor func() interface{}

// UnpackFunc must return a pointer to a struct and the resource id
type UnpackFunc func(s *schema.ResourceData) (interface{}, string, error)

type PackFunc func(repo interface{}, d *schema.ResourceData) error

func mkRepoCreate(unpack UnpackFunc, read ReadFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		repo, key, err := unpack(d)
		if err != nil {
			return err
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().AddRetryCondition(func(response *resty.Response, _r error) bool {
			return regexp.MustCompile(".*Could not merge and save new descriptor.*").MatchString(string(response.Body()[:]))
		}).SetBody(repo).Put(repositoriesEndpoint + key)

		if err != nil {
			return err
		}
		d.SetId(key)
		return read(d, m)
	}
}

func mkRepoRead(pack PackFunc, construct Constructor) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		repo := construct()
		// repo must be a pointer
		resp, err := m.(*resty.Client).R().SetResult(repo).Get(repositoriesEndpoint + d.Id())

		if err != nil {
			if resp != nil && (resp.StatusCode() == http.StatusNotFound) {
				d.SetId("")
				return nil
			}
			return err
		}
		return pack(repo, d)
	}
}

func mkRepoUpdate(unpack UnpackFunc, read ReadFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		repo, key, err := unpack(d)
		if err != nil {
			return err
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().SetBody(repo).Post(repositoriesEndpoint + d.Id())
		if err != nil {
			return err
		}

		d.SetId(key)
		return read(d, m)
	}
}

func deleteRepo(d *schema.ResourceData, m interface{}) error {
	resp, err := m.(*resty.Client).R().Delete(repositoriesEndpoint + d.Id())

	if err != nil && (resp != nil && resp.StatusCode() == http.StatusNotFound) {
		d.SetId("")
		return nil
	}
	return err
}

var neverRetry = func(response *resty.Response, err error) bool {
	return false
}

var retry400 = func(response *resty.Response, err error) bool {
	return response.StatusCode() == 400
}

func checkRepo(id string, m interface{}, retryCond resty.RetryConditionFunc) (bool, error) {
	_, err := m.(*resty.Client).R().AddRetryCondition(retryCond).Head(repositoriesEndpoint + id)
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return err == nil, err
}

func repoExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return checkRepo(d.Id(), m, retry400)
}

var repoTypeValidator = validation.StringInSlice(repoTypesSupported, false)

var repoKeyValidator = validation.All(
	validation.StringDoesNotMatch(regexp.MustCompile("^[0-9].*"), "repo key cannot start with a number"),
	validation.StringDoesNotContainAny(" !@#$%^&*()_+={}[]:;<>,/?~`|\\"),
)

var repoTypesSupported = []string{
	"alpine",
	"bower",
	//"cargo", // not supported
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
var baseLocalRepoSchema = map[string]*schema.Schema{
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
		Computed:     true,
		ValidateFunc: repoTypeValidator,
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
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"blacked_out": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},

	"xray_index": {
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
	"archive_browsing_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"optional_index_compression_formats": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Set:      schema.HashString,
		Optional: true,
	},
	"download_direct": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"block_pushing_schema1": {
		Type:     schema.TypeBool,
		Optional: true,
	},
}
var baseRemoteSchema = map[string]*schema.Schema{
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
	"xray_index": {
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
	"failed_retrieval_cache_period_secs": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
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
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
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
	"bypass_head_requests": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"client_tls_certificate": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"block_push_schema1": {
		Type:     schema.TypeBool,
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
}
var baseVirtualRepoSchema = map[string]*schema.Schema{
	"key": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
	"package_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repoTypeValidator,
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
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"repositories": {
		Type:     schema.TypeList,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Required: true,
	},

	"artifactory_requests_can_retrieve_remote_artifacts": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"default_deployment_repo": {
		Type:     schema.TypeString,
		Optional: true,
	},
}

func packBaseRemoteRepo(d *schema.ResourceData, repo services.RemoteRepositoryBaseParams) Lens {
	setValue := mkLens(d)
	setValue("key", repo.Key)
	setValue("package_type", repo.PackageType)
	setValue("url", repo.Url)
	setValue("username", repo.Username)

	setValue("proxy", repo.Proxy)
	setValue("description", repo.Description)
	setValue("notes", repo.Notes)
	setValue("includes_pattern", repo.IncludesPattern)
	setValue("excludes_pattern", repo.ExcludesPattern)
	setValue("repo_layout_ref", repo.RepoLayoutRef)
	setValue("hard_fail", *repo.HardFail)
	setValue("offline", *repo.Offline)
	setValue("blacked_out", *repo.BlackedOut)
	setValue("xray_index", *repo.XrayIndex)
	setValue("store_artifacts_locally", *repo.StoreArtifactsLocally)
	setValue("socket_timeout_millis", repo.SocketTimeoutMillis)
	setValue("local_address", repo.LocalAddress)
	setValue("retrieval_cache_period_seconds", repo.RetrievalCachePeriodSecs)
	setValue("failed_retrieval_cache_period_secs", repo.FailedRetrievalCachePeriodSecs)
	setValue("missed_cache_period_seconds", repo.MissedRetrievalCachePeriodSecs)
	setValue("unused_artifacts_cleanup_period_hours", repo.UnusedArtifactsCleanupPeriodHours)
	setValue("assumed_offline_period_secs", repo.AssumedOfflinePeriodSecs)
	setValue("share_configuration", *repo.ShareConfiguration)
	setValue("synchronize_properties", *repo.SynchronizeProperties)
	setValue("block_mismatching_mime_types", *repo.BlockMismatchingMimeTypes)
	setValue("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(repo.PropertySets)))
	setValue("allow_any_host_auth", *repo.AllowAnyHostAuth)
	setValue("enable_cookie_management", *repo.EnableCookieManagement)
	setValue("bypass_head_requests", *repo.BypassHeadRequests)
	setValue("client_tls_certificate", repo.ClientTlsCertificate)
	setValue("block_push_schema1", *repo.BlockPushingSchema1)

	if repo.ContentSynchronisation != nil {
		setValue("content_synchronisation", []interface{}{
			map[string]bool{
				"enabled": repo.ContentSynchronisation.Enabled,
			},
		})
	}
	return setValue
}
func unpackBaseRemoteRepo(s *schema.ResourceData) services.RemoteRepositoryBaseParams {
	d := &ResourceData{s}

	repo := services.RemoteRepositoryBaseParams{
		Rclass:                            "remote",
		Key:                               d.getString("key", false),
		PackageType:                       d.getString("package_type", true),
		Url:                               d.getString("url", false),
		Username:                          d.getString("username", true),
		Password:                          d.getString("password", true),
		Proxy:                             d.getString("proxy", true),
		Description:                       d.getString("description", true),
		Notes:                             d.getString("notes", true),
		IncludesPattern:                   d.getString("includes_pattern", true),
		ExcludesPattern:                   d.getString("excludes_pattern", true),
		RepoLayoutRef:                     d.getString("repo_layout_ref", true),
		HardFail:                          d.getBoolRef("hard_fail", true),
		Offline:                           d.getBoolRef("offline", true),
		BlackedOut:                        d.getBoolRef("blacked_out", true),
		XrayIndex:                         d.getBoolRef("xray_index", true),
		StoreArtifactsLocally:             d.getBoolRef("store_artifacts_locally", true),
		SocketTimeoutMillis:               d.getInt("socket_timeout_millis", true),
		LocalAddress:                      d.getString("local_address", true),
		RetrievalCachePeriodSecs:          d.getInt("retrieval_cache_period_seconds", true),
		FailedRetrievalCachePeriodSecs:    d.getInt("failed_retrieval_cache_period_secs", true),
		MissedRetrievalCachePeriodSecs:    d.getInt("missed_cache_period_seconds", true),
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
		BlockPushingSchema1:               d.getBoolRef("block_push_schema1", true),
	}

	if v, ok := d.GetOk("content_synchronisation"); ok {
		contentSynchronisationConfig := v.([]interface{})[0].(map[string]interface{})
		enabled := contentSynchronisationConfig["enabled"].(bool)
		repo.ContentSynchronisation = &services.ContentSynchronisation{
			Enabled: enabled,
		}
	}
	return repo
}

func unpackBaseVirtRepo(s *schema.ResourceData) services.VirtualRepositoryBaseParams {
	d := &ResourceData{s}

	return services.VirtualRepositoryBaseParams{
		Key:             d.getString("key", false),
		Rclass:          "virtual",
		PackageType:     d.getString("package_type", false),
		IncludesPattern: d.getString("includes_pattern", false),
		ExcludesPattern: d.getString("excludes_pattern", false),
		RepoLayoutRef:   d.getString("repo_layout_ref", false),
		ArtifactoryRequestsCanRetrieveRemoteArtifacts: d.getBoolRef("artifactory_requests_can_retrieve_remote_artifacts", false),
		Repositories:          d.getList("repositories"),
		Description:           d.getString("description", false),
		Notes:                 d.getString("notes", false),
		DefaultDeploymentRepo: d.getString("default_deployment_repo", false),
	}
}

func packBaseVirtRepo(d *schema.ResourceData, repo services.VirtualRepositoryBaseParams) Lens {
	setValue := mkLens(d)

	setValue("key", repo.Key)
	setValue("package_type", repo.PackageType)
	setValue("description", repo.Description)
	setValue("notes", repo.Notes)
	setValue("includes_pattern", repo.IncludesPattern)
	setValue("excludes_pattern", repo.ExcludesPattern)
	setValue("repo_layout_ref", repo.RepoLayoutRef)
	setValue("artifactory_requests_can_retrieve_remote_artifacts", *repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	setValue("default_deployment_repo", repo.DefaultDeploymentRepo)
	setValue("repositories", repo.Repositories)
	return setValue
}
