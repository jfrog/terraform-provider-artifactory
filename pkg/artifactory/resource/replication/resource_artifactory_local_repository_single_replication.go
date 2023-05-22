package replication

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"github.com/jfrog/terraform-provider-shared/client"
)

type localSingleReplicationBody struct {
	URL                             string `json:"url"`
	SocketTimeoutMillis             int    `json:"socketTimeoutMillis"`
	Username                        string `json:"username"`
	Password                        string `json:"password"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	SyncStatistics                  bool   `json:"syncStatistics"`
	RepoKey                         string `json:"repoKey"`
	IncludePathPrefixPattern        string `json:"includePathPrefixPattern"`
	ExcludePathPrefixPattern        string `json:"excludePathPrefixPattern"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
}

type getLocalSingleReplicationBody struct {
	localSingleReplicationBody
	ProxyRef       string `json:"proxyRef"`
	ReplicationKey string `json:"replicationKey"`
}

type updateLocalSingleReplicationBody struct {
	localSingleReplicationBody
	Proxy string `json:"proxy"`
}

var localSingleReplicationSchema = map[string]*schema.Schema{
	"repo_key": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Repository name.",
	},
	"cron_exp": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "The Cron expression that determines when the next replication will be triggered.",
	},
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
	},
	"url": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
		Description:      "The URL of the target local repository on a remote Artifactory server. Use the format `https://<artifactory_url>/artifactory/<repository_name>`.",
	},
	"socket_timeout_millis": {
		Type:             schema.TypeInt,
		Optional:         true,
		Default:          15000,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		Description:      "The network timeout in milliseconds to use for remote operations.",
	},
	"username": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "The HTTP authentication username.",
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Use either the HTTP authentication password or identity token.",
	},
	"sync_deletes": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata). Note that enabling this option, will delete artifacts on the target that do not exist in the source repository. Default value is `false`.",
	},
	"sync_properties": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "When set, the task also synchronizes the properties of replicated artifacts. Default value is `true`",
	},
	"sync_statistics": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, the task also synchronizes artifact download statistics. Set to avoid inadvertent cleanup at the target instance when setting up replication for disaster recovery. Default value is `false`",
	},
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "When set, enables replication of this repository to the target specified in `url` attribute. Default value is `true`.",
	},
	"include_path_prefix_pattern": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"exclude_path_prefix_pattern": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.",
	},
	"proxy": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "A proxy configuration to use when communicating with the remote instance.",
	},
	"replication_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Replication ID. The ID is known only after the replication is created, for this reason it's `Computed` and can not be set by the user in HCL.",
	},
	"check_binary_existence_in_filestore": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: "Enabling the `check_binary_existence_in_filestore` flag requires an Enterprise+ license. When true, enables distributed checksum storage. For more information, see " +
			"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
	},
}

func unpackLocalSingleReplication(s *schema.ResourceData) updateLocalSingleReplicationBody {
	d := &utilsdk.ResourceData{ResourceData: s}

	return updateLocalSingleReplicationBody{
		localSingleReplicationBody: localSingleReplicationBody{
			URL:                             d.GetString("url", false),
			SocketTimeoutMillis:             d.GetInt("socket_timeout_millis", false),
			Username:                        d.GetString("username", false),
			Password:                        d.GetString("password", false),
			EnableEventReplication:          d.GetBool("enable_event_replication", false),
			Enabled:                         d.GetBool("enabled", false),
			CronExp:                         d.GetString("cron_exp", false),
			SyncDeletes:                     d.GetBool("sync_deletes", false),
			SyncProperties:                  d.GetBool("sync_properties", false),
			SyncStatistics:                  d.GetBool("sync_statistics", false),
			RepoKey:                         d.GetString("repo_key", false),
			IncludePathPrefixPattern:        d.GetString("include_path_prefix_pattern", false),
			ExcludePathPrefixPattern:        d.GetString("exclude_path_prefix_pattern", false),
			CheckBinaryExistenceInFilestore: d.GetBool("check_binary_existence_in_filestore", false),
		},
		Proxy: d.GetString("proxy", false),
	}
}

func packLocalSingleReplication(singleLocalReplication *getLocalSingleReplicationBody, d *schema.ResourceData) diag.Diagnostics {

	var errors []error
	setValue := utilsdk.MkLens(d)
	setValue("url", singleLocalReplication.URL)
	setValue("socket_timeout_millis", singleLocalReplication.SocketTimeoutMillis)
	setValue("username", singleLocalReplication.Username)
	setValue("enable_event_replication", singleLocalReplication.EnableEventReplication)
	setValue("enabled", singleLocalReplication.Enabled)
	setValue("cron_exp", singleLocalReplication.CronExp)
	setValue("sync_deletes", singleLocalReplication.SyncDeletes)
	setValue("sync_properties", singleLocalReplication.SyncProperties)
	setValue("sync_statistics", singleLocalReplication.SyncStatistics)
	setValue("repo_key", singleLocalReplication.RepoKey)
	setValue("include_path_prefix_pattern", singleLocalReplication.IncludePathPrefixPattern)
	setValue("exclude_path_prefix_pattern", singleLocalReplication.ExcludePathPrefixPattern)
	setValue("proxy", singleLocalReplication.ProxyRef)
	setValue("replication_key", singleLocalReplication.ReplicationKey)
	errors = setValue("check_binary_existence_in_filestore", singleLocalReplication.CheckBinaryExistenceInFilestore)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourceLocalSingleReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackLocalSingleReplication(d)

	if verified, err := verifyRepoRclass(pushReplication.RepoKey, "local", m); !verified {
		return diag.Errorf("source repository rclass is not local, only remote repositories are supported by this resource %v", err)
	}
	_, err := m.(utilsdk.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		Put(EndpointPath + pushReplication.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pushReplication.RepoKey)
	return resourceLocalSingleReplicationRead(ctx, d, m)
}

func resourceLocalSingleReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(utilsdk.ProvderMetadata).Client
	var replicationInterface interface{}

	resp, err := c.R().SetResult(&replicationInterface).Get(EndpointPath + d.Id())

	if err != nil {
		if resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	replicationList, ok := replicationInterface.([]interface{})
	if ok {
		jsonData, _ := json.Marshal(replicationList)
		var replications []getLocalSingleReplicationBody
		json.Unmarshal(jsonData, &replications)

		return packLocalSingleReplication(&replications[0], d)
	}

	replicationObj, ok := replicationInterface.(interface{})
	if ok {
		jsonData, _ := json.Marshal(replicationObj)
		var replication getLocalSingleReplicationBody
		json.Unmarshal(jsonData, &replication)

		return packLocalSingleReplication(&replication, d)
	}

	return diag.Errorf("Error during converting the API call payload.")
}

func resourceLocalSingleReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackLocalSingleReplication(d)

	if verified, err := verifyRepoRclass(pushReplication.RepoKey, "local", m); !verified {
		return diag.Errorf("source repository rclass is not local, only remote repositories are supported by this resource %v", err)
	}
	_, err := m.(utilsdk.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		AddRetryCondition(client.RetryOnMergeError).
		Post(EndpointPath + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceLocalSingleReplicationRead(ctx, d, m)
}

func ResourceArtifactoryLocalRepositorySingleReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLocalSingleReplicationCreate,
		ReadContext:   resourceLocalSingleReplicationRead,
		UpdateContext: resourceLocalSingleReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "Add or replace a single replication configuration for given repository key. Supported by local repositories. Artifactory Pro license is required.",
		Schema:      localSingleReplicationSchema,
	}
}
