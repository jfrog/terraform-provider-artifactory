package artifactory

import (
	"context"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
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
				Default:  true,
			},
			"replications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"socket_timeout_millis": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  15000,
						},
						"username": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"password": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
							StateFunc: GetMD5Hash,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"sync_deletes": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"sync_properties": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"sync_statistics": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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

func unmarshalReplicationConfig(s *schema.ResourceData) *artifactory.ReplicationConfig {
	d := &ResourceData{s}
	replicationConfig := new(artifactory.ReplicationConfig)

	repo := d.GetStringRef("repo_key")

	if v, ok := d.GetOkExists("replications"); ok {
		arr := v.([]interface{})

		tmp := make([]artifactory.SingleReplicationConfig, 0, len(arr))
		replicationConfig.Replications = &tmp

		for i, o := range arr {
			if i == 0 {
				replicationConfig.RepoKey = repo
				replicationConfig.CronExp = d.GetStringRef("cron_exp")
				replicationConfig.EnableEventReplication = d.GetBoolRef("enable_event_replication")
			}

			m := o.(map[string]interface{})

			var replication artifactory.SingleReplicationConfig

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

func marshalReplicationConfig(replicationConfig *artifactory.ReplicationConfig, d *schema.ResourceData) {
	d.Set("repo_key", replicationConfig.RepoKey)
	d.Set("cron_exp", replicationConfig.CronExp)
	d.Set("enable_event_replication", replicationConfig.EnableEventReplication)

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
				replication["password"] = GetMD5Hash(*repo.Password)
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
		d.Set("replications", replications)
	}
}

func resourceReplicationConfigCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	replicationConfig := unmarshalReplicationConfig(d)

	_, err := c.Artifacts.SetRepositoryReplicationConfig(context.Background(), *replicationConfig.RepoKey, replicationConfig)
	if err != nil {
		return err
	}

	d.SetId(*replicationConfig.RepoKey)
	return resourceReplicationConfigRead(d, m)
}

func resourceReplicationConfigRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	replicationConfig, _, err := c.Artifacts.GetRepositoryReplicationConfig(context.Background(), d.Id())

	if err != nil {
		return err
	}

	marshalReplicationConfig(replicationConfig, d)
	return nil
}

func resourceReplicationConfigUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	replicationConfig := unmarshalReplicationConfig(d)
	_, err := c.Artifacts.UpdateRepositoryReplicationConfig(context.Background(), d.Id(), replicationConfig)
	if err != nil {
		return err
	}

	d.SetId(*replicationConfig.RepoKey)

	return resourceReplicationConfigRead(d, m)
}

func resourceReplicationConfigDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)
	replicationConfig := unmarshalReplicationConfig(d)
	_, err := c.Artifacts.DeleteRepositoryReplicationConfig(context.Background(), *replicationConfig.RepoKey)
	return err
}

func resourceReplicationConfigExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*artifactory.Client)

	replicationName := d.Id()
	_, resp, err := c.Artifacts.GetRepositoryReplicationConfig(context.Background(), replicationName)

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
