package artifactory

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

func resourceArtifactoryPullReplicationConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePullReplicationConfigCreate,
		ReadContext:   resourcePullReplicationConfigRead,
		UpdateContext: resourcePullReplicationConfigUpdate,
		DeleteContext: resourceReplicationConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      mergeSchema(replicationSchemaCommon, replicationSchema),
		Description: "Used for configuring pull replication on remote repos.",
	}
}

func unpackPullReplicationConfig(s *schema.ResourceData) *utils.ReplicationBody {
	d := &ResourceData{s}
	replicationConfig := new(utils.ReplicationBody)

	replicationConfig.RepoKey = d.getString("repo_key", false)
	replicationConfig.CronExp = d.getString("cron_exp", false)
	replicationConfig.EnableEventReplication = d.getBool("enable_event_replication", false)
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

	setValue("enabled", config.Enabled)
	setValue("sync_deletes", config.SyncDeletes)
	setValue("sync_properties", config.SyncProperties)

	errors := setValue("path_prefix", config.PathPrefix)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}
func resourcePullReplicationConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackPullReplicationConfig(d)
	// The password is sent clear
	_, err := m.(*resty.Client).R().SetBody(replicationConfig).Put(replicationEndpoint + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)
	return resourcePullReplicationConfigRead(ctx, d, m)
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
}

func resourcePullReplicationConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var result interface{}

	resp, err := m.(*resty.Client).R().SetResult(&result).Get(replicationEndpoint + d.Id())
	// password comes back scrambled
	if err != nil {
		return diag.FromErr(err)
	}

	final := PullReplication{}
	err = json.Unmarshal(resp.Body(), &final)
	if err != nil {
		return diag.FromErr(err)
	}
	return packPullReplicationBody(final, d)
}

func resourcePullReplicationConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackPullReplicationConfig(d)
	_, err := m.(*resty.Client).R().SetBody(replicationConfig).Post(replicationEndpoint + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)

	return resourcePullReplicationConfigRead(ctx, d, m)
}
