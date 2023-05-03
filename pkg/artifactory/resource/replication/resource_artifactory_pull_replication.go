package replication

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"github.com/jfrog/terraform-provider-shared/client"
)

// PullReplication this is the structure for a PULL replication on a remote repo
type PullReplication struct {
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	PathPrefix                      string `json:"pathPrefix"`
	RepoKey                         string `json:"repoKey"`
	ReplicationKey                  string `json:"replicationKey"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	Username                        string `json:"username"`
	Password                        string `json:"password"`
	URL                             string `json:"url"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
}

var pullReplicationSchema = map[string]*schema.Schema{
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
	"url": {
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		RequiredWith:     []string{"username", "password"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
		Description:      "URL for local repository replication. Required for local repository, but not needed for remote repository.",
	},
	"socket_timeout_millis": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
	},
	"username": {
		Type:             schema.TypeString,
		Optional:         true,
		RequiredWith:     []string{"url", "password"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Username for local repository replication. Required for local repository, but not needed for remote repository.",
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		RequiredWith:     []string{"url", "username"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Password for local repository replication. Required for local repository, but not needed for remote repository.",
	},
	"enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"sync_deletes": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"sync_properties": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"sync_statistics": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"path_prefix": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"proxy": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Proxy key from Artifactory Proxies setting",
	},
	"check_binary_existence_in_filestore": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: "When true, enables distributed checksum storage. For more information, see " +
			"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
	},
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
	},
}

func unpackPullReplication(s *schema.ResourceData) *ReplicationBody {
	d := &utilsdk.ResourceData{ResourceData: s}
	replicationConfig := new(ReplicationBody)

	replicationConfig.RepoKey = d.GetString("repo_key", false)
	replicationConfig.CronExp = d.GetString("cron_exp", false)
	replicationConfig.EnableEventReplication = d.GetBool("enable_event_replication", false)
	replicationConfig.URL = d.GetString("url", false)
	replicationConfig.Username = d.GetString("username", false)
	replicationConfig.Password = d.GetString("password", false)
	replicationConfig.Enabled = d.GetBool("enabled", false)
	replicationConfig.SyncDeletes = d.GetBool("sync_deletes", false)
	replicationConfig.SyncProperties = d.GetBool("sync_properties", false)
	replicationConfig.SyncStatistics = d.GetBool("sync_statistics", false)
	replicationConfig.PathPrefix = d.GetString("path_prefix", false)
	replicationConfig.CheckBinaryExistenceInFilestore = d.GetBool("check_binary_existence_in_filestore", false)

	return replicationConfig
}

func packPullReplication(config PullReplication, d *schema.ResourceData) diag.Diagnostics {
	setValue := utilsdk.MkLens(d)

	setValue("repo_key", config.RepoKey)
	setValue("cron_exp", config.CronExp)
	setValue("enable_event_replication", config.EnableEventReplication)
	setValue("url", config.URL)
	setValue("username", config.Username)
	setValue("enabled", config.Enabled)
	setValue("sync_deletes", config.SyncDeletes)
	setValue("sync_properties", config.SyncProperties)
	setValue("check_binary_existence_in_filestore", config.CheckBinaryExistenceInFilestore)

	errors := setValue("path_prefix", config.PathPrefix)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourcePullReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackPullReplication(d)
	// The password is sent clear
	_, err := m.(utilsdk.ProvderMetadata).Client.R().
		SetBody(replicationConfig).
		AddRetryCondition(client.RetryOnMergeError).
		Put(EndpointPath + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)
	return resourcePullReplicationRead(ctx, d, m)
}

func resourcePullReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var result interface{}

	resp, err := m.(utilsdk.ProvderMetadata).Client.R().SetResult(&result).Get(EndpointPath + d.Id())
	// password comes back scrambled
	if err != nil {
		return diag.FromErr(err)
	}

	switch result.(type) {
	case []interface{}:
		if len(result.([]interface{})) > 1 {
			return diag.Errorf("received more than one replication payload. expect only one in array")
		}
		var final []PullReplication
		err = json.Unmarshal(resp.Body(), &final)
		if err != nil {
			return diag.FromErr(err)
		}
		return packPullReplication(final[0], d)
	default:
		final := PullReplication{}
		err = json.Unmarshal(resp.Body(), &final)
		if err != nil {
			return diag.FromErr(err)
		}
		return packPullReplication(final, d)
	}
}

func resourcePullReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackPullReplication(d)
	_, err := m.(utilsdk.ProvderMetadata).Client.R().
		SetBody(replicationConfig).
		AddRetryCondition(client.RetryOnMergeError).
		Post(EndpointPath + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)

	return resourcePullReplicationRead(ctx, d, m)
}

func ResourceArtifactoryPullReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePullReplicationCreate,
		ReadContext:   resourcePullReplicationRead,
		UpdateContext: resourcePullReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:             pullReplicationSchema,
		Description:        "Used for configuring pull replication on local or remote repos.",
		DeprecationMessage: "This resource is replaced by `artifactory_remote_repository_replication` for clarity. Please consider the migration.",
	}
}
