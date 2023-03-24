package replication

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactorySingleReplicationConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSingleReplicationConfigCreate,
		ReadContext:   resourceSingleReplicationConfigRead,
		UpdateContext: resourceSingleReplicationConfigUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: util.MergeMaps(replicationSchemaCommon, replicationSchema),
		Description: "Used for configuring replications on repos. However, the TCL only makes " +
			"good sense for local repo replication (PUSH) and not remote (PULL).",
		DeprecationMessage: "This resource has been deprecated in favour of the more explicitly name" +
			"artifactory_pull_replication resource.",
	}
}

func unpackSingleReplicationConfig(s *schema.ResourceData) *updateReplicationBody {
	d := &util.ResourceData{ResourceData: s}
	replicationConfig := new(updateReplicationBody)

	replicationConfig.RepoKey = d.GetString("repo_key", false)
	replicationConfig.CronExp = d.GetString("cron_exp", false)
	replicationConfig.EnableEventReplication = d.GetBool("enable_event_replication", false)
	replicationConfig.URL = d.GetString("url", false)
	replicationConfig.SocketTimeoutMillis = d.GetInt("socket_timeout_millis", false)
	replicationConfig.Username = d.GetString("username", false)
	replicationConfig.Enabled = d.GetBool("enabled", false)
	replicationConfig.SyncDeletes = d.GetBool("sync_deletes", false)
	replicationConfig.SyncProperties = d.GetBool("sync_properties", false)
	replicationConfig.SyncStatistics = d.GetBool("sync_statistics", false)
	replicationConfig.PathPrefix = d.GetString("path_prefix", false)
	replicationConfig.Proxy = repository.HandleResetWithNonExistentValue(d, "proxy")
	replicationConfig.Password = d.GetString("password", false)

	return replicationConfig
}

func packPushReplicationBody(config getReplicationBody, d *schema.ResourceData) diag.Diagnostics {
	setValue := util.MkLens(d)

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

	setValue("path_prefix", config.PathPrefix)

	errors := setValue("proxy", config.ProxyRef)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func packPullReplicationBody(config PullReplication, d *schema.ResourceData) diag.Diagnostics {
	setValue := util.MkLens(d)

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
	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(replicationConfig).
		AddRetryCondition(client.RetryOnMergeError).
		Put(EndpointPath + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)
	return resourceSingleReplicationConfigRead(ctx, d, m)
}

func resourceSingleReplicationConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// this endpoint serves for both PULL type replications (remote repo) and PUSH type replications
	// (local repos). In the case of a remote (pull), it's a singular object. In case of local (push), it's an array
	// If we query replications/ it will tell us which is which, but the direct query does not.
	// I don't like the idea of interrogating the data type but I also don't like having to make 2 api calls either
	// Frankly, the whole api sucks. We are going to reimplement it as atlassian did, but really, this needed to be
	// an entirely different resource because values like "url" are never available after submit.
	var result interface{}

	resp, err := m.(util.ProvderMetadata).Client.R().SetResult(&result).Get(EndpointPath + d.Id())
	// password comes back scrambled
	if err != nil {
		if resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	switch result.(type) {
	case []interface{}:
		if len(result.([]interface{})) > 1 {
			return diag.Errorf("resource_single_replication_config does not support multiple replication config on a repo. Use resource_artifactory_replication_config instead")
		}
		var final []getReplicationBody
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
	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(replicationConfig).
		AddRetryCondition(client.RetryOnMergeError).
		Post(EndpointPath + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)

	return resourceSingleReplicationConfigRead(ctx, d, m)
}
