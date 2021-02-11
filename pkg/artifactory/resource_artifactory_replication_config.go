package artifactory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceArtifactoryReplicationConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicationConfigCreate,
		Read:   resourceReplicationConfigRead,
		Update: resourceReplicationConfigUpdate,
		Delete: resourceReplicationConfigDelete,
		Exists: resourceReplicationConfigExists,

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
			"replications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
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
				},
			},
		},
	}
}

func unpackReplicationConfig(s *schema.ResourceData) *v1.ReplicationConfig {
	d := &ResourceData{s}
	replicationConfig := new(v1.ReplicationConfig)

	repo := d.getStringRef("repo_key", false)

	if v, ok := d.GetOkExists("replications"); ok {
		arr := v.([]interface{})

		tmp := make([]v1.SingleReplicationConfig, 0, len(arr))
		replicationConfig.Replications = &tmp

		for i, o := range arr {
			if i == 0 {
				replicationConfig.RepoKey = repo
				replicationConfig.CronExp = d.getStringRef("cron_exp", false)
				replicationConfig.EnableEventReplication = d.getBoolRef("enable_event_replication", false)
			}

			m := o.(map[string]interface{})

			var replication v1.SingleReplicationConfig

			replication.RepoKey = repo

			if v, ok := m["url"]; ok {
				replication.URL = artifactory.String(v.(string))
			}

			if v, ok := m["socket_timeout_millis"]; ok {
				replication.SocketTimeoutMillis = artifactory.Int(v.(int))
			}

			if v, ok := m["username"]; ok {
				replication.Username = artifactory.String(v.(string))
			}

			if v, ok := m["enabled"]; ok {
				replication.Enabled = artifactory.Bool(v.(bool))
			}

			if v, ok := m["sync_deletes"]; ok {
				replication.SyncDeletes = artifactory.Bool(v.(bool))
			}

			if v, ok := m["sync_properties"]; ok {
				replication.SyncProperties = artifactory.Bool(v.(bool))
			}

			if v, ok := m["sync_statistics"]; ok {
				replication.SyncStatistics = artifactory.Bool(v.(bool))
			}

			if prefix, ok := m["path_prefix"]; ok {
				replication.PathPrefix = artifactory.String(prefix.(string))
			}

			if pass, ok := m["password"]; ok {
				replication.Password = artifactory.String(pass.(string))
			}

			*replicationConfig.Replications = append(*replicationConfig.Replications, replication)
		}
	}

	return replicationConfig
}

func packReplicationConfig(replicationConfig *v1.ReplicationConfig, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("repo_key", replicationConfig.RepoKey))
	logErr(d.Set("cron_exp", replicationConfig.CronExp))
	logErr(d.Set("enable_event_replication", replicationConfig.EnableEventReplication))

	if replicationConfig.Replications != nil {
		var replications []map[string]interface{}
		for _, repo := range *replicationConfig.Replications {
			replication := make(map[string]interface{})

			if repo.URL != nil {
				replication["url"] = *repo.URL
			}

			if repo.SocketTimeoutMillis != nil {
				replication["socket_timeout_millis"] = *repo.SocketTimeoutMillis
			}

			if repo.Username != nil {
				replication["username"] = *repo.Username
			}

			if repo.Password != nil {
				replication["password"] = getMD5Hash(*repo.Password)
			}

			if repo.Enabled != nil {
				replication["enabled"] = *repo.Enabled
			}

			if repo.SyncDeletes != nil {
				replication["sync_deletes"] = *repo.SyncDeletes
			}

			if repo.SyncProperties != nil {
				replication["sync_properties"] = *repo.SyncProperties
			}

			if repo.SyncStatistics != nil {
				replication["sync_statistics"] = *repo.SyncStatistics
			}

			if repo.PathPrefix != nil {
				replication["path_prefix"] = *repo.PathPrefix
			}

			replications = append(replications, replication)
		}

		logErr(d.Set("replications", replications))
	}

	if hasErr {
		return fmt.Errorf("failed to pack replication config")
	}

	return nil
}

func resourceReplicationConfigCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	replicationConfig := unpackReplicationConfig(d)

	_, err := c.V1.Artifacts.SetRepositoryReplicationConfig(context.Background(), *replicationConfig.RepoKey, replicationConfig)
	if err != nil {
		return err
	}

	d.SetId(*replicationConfig.RepoKey)
	return resourceReplicationConfigRead(d, m)
}

func resourceReplicationConfigRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	replicationConfig, _, err := c.V1.Artifacts.GetRepositoryReplicationConfig(context.Background(), d.Id())

	if err != nil {
		return err
	}

	return packReplicationConfig(replicationConfig, d)
}

func resourceReplicationConfigUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	replicationConfig := unpackReplicationConfig(d)
	_, err := c.V1.Artifacts.UpdateRepositoryReplicationConfig(context.Background(), d.Id(), replicationConfig)
	if err != nil {
		return err
	}

	d.SetId(*replicationConfig.RepoKey)

	return resourceReplicationConfigRead(d, m)
}

func resourceReplicationConfigDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld
	replicationConfig := unpackReplicationConfig(d)
	_, err := c.V1.Artifacts.DeleteRepositoryReplicationConfig(context.Background(), *replicationConfig.RepoKey)
	return err
}

func resourceReplicationConfigExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*ArtClient).ArtOld

	replicationName := d.Id()
	_, resp, err := c.V1.Artifacts.GetRepositoryReplicationConfig(context.Background(), replicationName)

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
