package artifactory

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)

const replicationEndpoint = "artifactory/api/replications/"

func resourceArtifactorySingleReplicationConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSingleReplicationConfigCreate,
		ReadContext:   resourceSingleReplicationConfigRead,
		UpdateContext: resourceSingleReplicationConfigUpdate,
		DeleteContext: resourceReplicationConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: mergeSchema(replicationSchemaCommon, replicationSchema),
		Description: "Used for configuring replications on repos. However, the TCL only makes " +
			"good sense for remote (PULL) repo replication (PUSH) and not local (PUSH).",
		DeprecationMessage: "The APIs underpinning this resource support local and remote repository replication, " +
			"but their payloads are entirely different. You should only use this for remote repository replication.",
	}
}

func unpackSingleReplicationConfig(s *schema.ResourceData) *utils.ReplicationBody {
	d := &ResourceData{s}
	replicationConfig := new(utils.ReplicationBody)

	replicationConfig.RepoKey = d.getString("repo_key", false)
	replicationConfig.CronExp = d.getString("cron_exp", false)
	replicationConfig.EnableEventReplication = d.getBool("enable_event_replication", false)
	replicationConfig.URL = d.getString("url", false)
	replicationConfig.SocketTimeoutMillis = d.getInt("socket_timeout_millis", false)
	replicationConfig.Username = d.getString("username", false)
	replicationConfig.Enabled = d.getBool("enabled", false)
	replicationConfig.SyncDeletes = d.getBool("sync_deletes", false)
	replicationConfig.SyncProperties = d.getBool("sync_properties", false)
	replicationConfig.SyncStatistics = d.getBool("sync_statistics", false)
	replicationConfig.PathPrefix = d.getString("path_prefix", false)
	replicationConfig.Password = d.getString("password", false)

	return replicationConfig
}

func packPushReplicationBody(config utils.ReplicationBody, d *schema.ResourceData) diag.Diagnostics {
	setValue := mkLens(d)

	setValue("repo_key", config.RepoKey)
	setValue("cron_exp", config.CronExp)
	setValue("enable_event_replication", config.EnableEventReplication)

	setValue("url", config.URL)
	setValue("socket_timeout_millis", config.SocketTimeoutMillis)
	setValue("username", config.Username)
	// the password coming back from artifactory is already scrambled, and I don't know in what form.
	// password -> JE2fNsEThvb1buiH7h7S2RDsGWSdp2EcuG9Pky5AFyRMwE4UzG
	// Because it comes back scrambled, we can't/shouldn't touch it.
	setValue("password", config.Password)
	setValue("enabled", config.Enabled)
	setValue("sync_deletes", config.SyncDeletes)
	setValue("sync_properties", config.SyncProperties)
	setValue("sync_statistics", config.SyncStatistics)

	errors := setValue("path_prefix", config.PathPrefix)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
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
func resourceSingleReplicationConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackSingleReplicationConfig(d)
	// The password is sent clear
	_, err := m.(*resty.Client).R().SetBody(replicationConfig).Put(replicationEndpoint + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)
	return resourceSingleReplicationConfigRead(ctx, d, m)
}

// ReplicationSummary this is what you would get if you hit replications/
type ReplicationSummary struct {
	ReplicationType                 string `json:"replicationType"`
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	PathPrefix                      string `json:"pathPrefix"`
	RepoKey                         string `json:"repoKey"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	CheckBinaryExistenceInFileStore bool   `json:"checkBinaryExistenceInFilestore"`
	SyncStatistics                  bool   `json:"syncStatistics"`
}

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
	CheckBinaryExistenceInFileStore bool   `json:"checkBinaryExistenceInFilestore"`
}

func resourceSingleReplicationConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// this endpoint serves for both PULL type replications (remote repo) and PUSH type replications
	// (local repos). In the case of a remote (pull), it's a singular object. In case of local (push), it's an array
	// If we query replications/ it will tell us which is which, but the direct query does not.
	// I don't like the idea of interrogating the data type but I also don't like having to make 2 api calls either
	// Frankly, the whole api sucks. We are going to reimplement it as atlassian did, but really, this needed to be
	// an entirely different resource because values like "url" are never available after submit.
	var result interface{}

	resp, err := m.(*resty.Client).R().SetResult(&result).Get(replicationEndpoint + d.Id())
	// password comes back scrambled
	if err != nil {
		return diag.FromErr(err)
	}

	switch result.(type) {
	case []interface{}:
		if len(result.([]interface{})) > 1 {
			return diag.Errorf("resource_single_replication_config does not support multiple replication config on a repo. Use resource_artifactory_replication_config instead")
		}
		var final []utils.ReplicationBody
		err = json.Unmarshal(resp.Body(), &final)
		if err != nil {
			return diag.FromErr(err)
		}
		return packPushReplicationBody(final[0], d)
	default:
		final := PullReplication{}
		err = json.Unmarshal(resp.Body(), &final)
		if err != nil {
			return diag.FromErr(err)
		}
		return packPullReplicationBody(final, d)
	}
}

func resourceSingleReplicationConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackSingleReplicationConfig(d)
	_, err := m.(*resty.Client).R().SetBody(replicationConfig).Post(replicationEndpoint + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)

	return resourceSingleReplicationConfigRead(ctx, d, m)
}
