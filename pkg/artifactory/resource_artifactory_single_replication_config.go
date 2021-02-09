package artifactory

import (
	"context"
	"fmt"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"net/http"
)

func resourceArtifactorySingleReplicationConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceSingleReplicationConfigCreate,
		Read:   resourceSingleReplicationConfigRead,
		Update: resourceSingleReplicationConfigUpdate,
		Delete: resourceSingleReplicationConfigDelete,
		Exists: resourceSingleReplicationConfigExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"repo_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cron_exp": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enable_event_replication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"socket_timeout_millis": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
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
		},
	}
}

func unpackSingleReplicationConfig(s *schema.ResourceData) *v1.SingleReplicationConfig {
	d := &ResourceData{s}
	replicationConfig := new(v1.SingleReplicationConfig)

	replicationConfig.RepoKey = d.getStringRef("repo_key", false)
	replicationConfig.CronExp = d.getStringRef("cron_exp", false)
	replicationConfig.EnableEventReplication = d.getBoolRef("enable_event_replication", false)
	replicationConfig.URL = d.getStringRef("url", false)
	replicationConfig.SocketTimeoutMillis = d.getIntRef("socket_timeout_millis", false)
	replicationConfig.Username = d.getStringRef("username", false)
	replicationConfig.Enabled = d.getBoolRef("enabled", false)
	replicationConfig.SyncDeletes = d.getBoolRef("sync_deletes", false)
	replicationConfig.SyncProperties = d.getBoolRef("sync_properties", false)
	replicationConfig.SyncStatistics = d.getBoolRef("sync_statistics", false)
	replicationConfig.PathPrefix = d.getStringRef("path_prefix", false)
	replicationConfig.Password = d.getStringRef("password", false)

	return replicationConfig
}

func packSingleReplicationConfig(replicationConfig *v1.ReplicationConfig, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("repo_key", replicationConfig.RepoKey))
	logErr(d.Set("cron_exp", replicationConfig.CronExp))
	logErr(d.Set("enable_event_replication", replicationConfig.EnableEventReplication))

	firstConfig := (*replicationConfig.Replications)[0]

	if firstConfig.URL != nil {
		logErr(d.Set("url", *firstConfig.URL))
	}

	if firstConfig.SocketTimeoutMillis != nil {
		logErr(d.Set("socket_timeout_millis", *firstConfig.SocketTimeoutMillis))
	}

	if firstConfig.Username != nil {
		logErr(d.Set("username", *firstConfig.Username))
	}

	if firstConfig.Password != nil {
		logErr(d.Set("password", getMD5Hash(*firstConfig.Password)))
	}

	if firstConfig.Enabled != nil {
		logErr(d.Set("enabled", *firstConfig.Enabled))
	}

	if firstConfig.SyncDeletes != nil {
		logErr(d.Set("sync_deletes", *firstConfig.SyncDeletes))
	}

	if firstConfig.SyncProperties != nil {
		logErr(d.Set("sync_properties", *firstConfig.SyncProperties))
	}

	if firstConfig.SyncStatistics != nil {
		logErr(d.Set("sync_statistics", *firstConfig.SyncStatistics))
	}

	if firstConfig.PathPrefix != nil {
		logErr(d.Set("path_prefix", *firstConfig.PathPrefix))
	}

	if hasErr {
		return fmt.Errorf("failed to pack replication config")
	}

	return nil
}

func resourceSingleReplicationConfigCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	replicationConfig := unpackSingleReplicationConfig(d)

	_, err := c.V1.Artifacts.SetSingleRepositoryReplicationConfig(context.Background(), *replicationConfig.RepoKey, replicationConfig)
	if err != nil {
		return err
	}

	d.SetId(*replicationConfig.RepoKey)
	return resourceSingleReplicationConfigRead(d, m)
}

func resourceSingleReplicationConfigRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	replicationConfig, _, err := c.V1.Artifacts.GetRepositoryReplicationConfig(context.Background(), d.Id())

	if err != nil {
		return err
	} else if len(*replicationConfig.Replications) > 1 {
		return fmt.Errorf("resource_single_replication_config does not support multiple replication config on a repo. Use resource_artifactory_replication_config instead")
	}

	return packSingleReplicationConfig(replicationConfig, d)
}

func resourceSingleReplicationConfigUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	replicationConfig := unpackSingleReplicationConfig(d)
	_, err := c.V1.Artifacts.UpdateSingleRepositoryReplicationConfig(context.Background(), d.Id(), replicationConfig)
	if err != nil {
		return err
	}

	d.SetId(*replicationConfig.RepoKey)

	return resourceSingleReplicationConfigRead(d, m)
}

func resourceSingleReplicationConfigDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld
	replicationConfig := unpackSingleReplicationConfig(d)
	_, err := c.V1.Artifacts.DeleteRepositoryReplicationConfig(context.Background(), *replicationConfig.RepoKey)
	return err
}

func resourceSingleReplicationConfigExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*ArtClient).ArtOld

	replicationName := d.Id()
	replicationConfig, resp, err := c.V1.Artifacts.GetRepositoryReplicationConfig(context.Background(), replicationName)

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	} else if len(*replicationConfig.Replications) > 1 {
		return false, fmt.Errorf("resource_single_replication_config does not support multiple replication config on a repo. Use resource_artifactory_replication_config instead")
	}

	return true, nil
}
