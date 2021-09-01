package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
)
const replicationEndpoint = "artifactory/api/replications/"

func resourceArtifactorySingleReplicationConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceSingleReplicationConfigCreate,
		Read:   resourceSingleReplicationConfigRead,
		Update: resourceSingleReplicationConfigUpdate,
		Delete: resourceReplicationConfigDelete,
		Exists: resourceReplicationConfigExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: mergeSchema(replicationSchemaCommon,replicationSchema),
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

func packSingleReplicationConfig(config *utils.ReplicationBody, d *schema.ResourceData) error {
	setValue := mkLens(d)

	setValue("repo_key", config.RepoKey)
	setValue("cron_exp", config.CronExp)
	setValue("enable_event_replication", config.EnableEventReplication)


	setValue("url", config.URL)
	setValue("socket_timeout_millis", config.SocketTimeoutMillis)
	setValue("username", config.Username)
	setValue("password", getMD5Hash(config.Password))
	setValue("enabled", config.Enabled)
	setValue("sync_deletes", config.SyncDeletes)
	setValue("sync_properties", config.SyncProperties)
	setValue("sync_statistics", config.SyncStatistics)

	errors := setValue("path_prefix", config.PathPrefix)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack replication config %q",errors)
	}

	return nil
}

func resourceSingleReplicationConfigCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	replicationConfig := unpackSingleReplicationConfig(d)

	_,err := client.R().SetBody(replicationConfig).Put(replicationEndpoint + replicationConfig.RepoKey)
	if err != nil {
		return err
	}

	d.SetId(replicationConfig.RepoKey)
	return resourceSingleReplicationConfigRead(d, m)
}

func resourceSingleReplicationConfigRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty
	replications := new([]utils.ReplicationBody)
	_, err := client.R().SetResult(replications).Get(replicationEndpoint + d.Id())

	if err != nil {
		return err
	}
	if len(*replications) > 1 {
		return fmt.Errorf("resource_single_replication_config does not support multiple replication config on a repo. Use resource_artifactory_replication_config instead")
	}
	return packSingleReplicationConfig(&(*replications)[0], d)
}

func resourceSingleReplicationConfigUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	replicationConfig := unpackSingleReplicationConfig(d)
	_, err := client.R().SetBody(replicationConfig).Post(replicationEndpoint + replicationConfig.RepoKey)
	if err != nil {
		return err
	}

	d.SetId(replicationConfig.RepoKey)

	return resourceSingleReplicationConfigRead(d, m)
}
