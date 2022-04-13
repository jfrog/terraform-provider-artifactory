package artifactory

import (
	"context"
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var pullReplicationSchema = map[string]*schema.Schema{
	"url": {
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		RequiredWith:     []string{"username", "password"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
		Description:      "(Optional) URL for local repository replication. Required for local repository, but not needed for remote repository.",
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
		Description:      "(Optional) Username for local repository replication. Required for local repository, but not needed for remote repository.",
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		RequiredWith:     []string{"url", "username"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "(Optional) Password for local repository replication. Required for local repository, but not needed for remote repository.",
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
		Type:     schema.TypeString,
		Optional: true,
		Description: "Proxy key from Artifactory Proxies setting",
	},
}

func resourceArtifactoryPullReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePullReplicationCreate,
		ReadContext:   resourcePullReplicationRead,
		UpdateContext: resourcePullReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      mergeSchema(replicationSchemaCommon, pullReplicationSchema),
		Description: "Used for configuring pull replication on local or remote repos.",
	}
}

func unpackPullReplication(s *schema.ResourceData) *ReplicationBody {
	d := &ResourceData{s}
	replicationConfig := new(ReplicationBody)

	replicationConfig.RepoKey = d.getString("repo_key", false)
	replicationConfig.CronExp = d.getString("cron_exp", false)
	replicationConfig.EnableEventReplication = d.getBool("enable_event_replication", false)
	replicationConfig.URL = d.getString("url", false)
	replicationConfig.Username = d.getString("username", false)
	replicationConfig.Password = d.getString("password", false)
	replicationConfig.Enabled = d.getBool("enabled", false)
	replicationConfig.SyncDeletes = d.getBool("sync_deletes", false)
	replicationConfig.SyncProperties = d.getBool("sync_properties", false)
	replicationConfig.SyncStatistics = d.getBool("sync_statistics", false)
	replicationConfig.PathPrefix = d.getString("path_prefix", false)

	return replicationConfig
}

func packPullReplicationBody(config PullReplication, d *schema.ResourceData) diag.Diagnostics {
	setValue := mkLens(d)

	setValue("repo_key", config.RepoKey)
	setValue("cron_exp", config.CronExp)
	setValue("enable_event_replication", config.EnableEventReplication)
	setValue("username", config.Username)
	setValue("enabled", config.Enabled)
	setValue("sync_deletes", config.SyncDeletes)
	setValue("sync_properties", config.SyncProperties)

	errors := setValue("path_prefix", config.PathPrefix)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourcePullReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackPullReplication(d)
	// The password is sent clear
	_, err := m.(*resty.Client).R().SetBody(replicationConfig).Put(replicationEndpointPath + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)
	return resourcePullReplicationRead(ctx, d, m)
}

// PullReplication this is the structure for a PULL replication on a remote repo
type PullReplication struct {
	Enabled                bool   `json:"enabled"`
	CronExp                string `json:"cronExp"`
	SyncDeletes            bool   `json:"syncDeletes"`
	SyncProperties         bool   `json:"syncProperties"`
	PathPrefix             string `json:"pathPrefix"`
	RepoKey                string `json:"repoKey"`
	ReplicationKey         string `json:"replicationKey"`
	EnableEventReplication bool   `json:"enableEventReplication"`
	Username               string `json:"username"`
	Password               string `json:"password"`
	URL                    string `json:"url"`
}

func resourcePullReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var result interface{}

	resp, err := m.(*resty.Client).R().SetResult(&result).Get(replicationEndpointPath + d.Id())
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
		return packPullReplicationBody(final[0], d)
	default:
		final := PullReplication{}
		err = json.Unmarshal(resp.Body(), &final)
		if err != nil {
			return diag.FromErr(err)
		}
		return packPullReplicationBody(final, d)
	}
}

func resourcePullReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackPullReplication(d)
	_, err := m.(*resty.Client).R().SetBody(replicationConfig).Post(replicationEndpointPath + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)

	return resourcePullReplicationRead(ctx, d, m)
}
