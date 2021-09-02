package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ReplicationConfig struct {
	RepoKey                string                  `json:"-"`
	CronExp                string                  `json:"cronExp,omitempty"`
	EnableEventReplication bool                    `json:"enableEventReplication,omitempty"`
	Replications           []utils.ReplicationBody `json:"replications,omitempty"`
}

var replicationSchemaCommon = map[string]*schema.Schema{
	"repo_key": {
		Type:     schema.TypeString,
		Required: true,
	},
	"cron_exp": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validateCron,
	},
	"enable_event_replication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
}

var repMultipleSchema = map[string]*schema.Schema{
	"replications": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: replicationSchema,
		},
	},
}
var replicationSchema = map[string]*schema.Schema{
	"url": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
	},
	"socket_timeout_millis": {
		Type:     schema.TypeInt,
		Optional: true,
		Computed: true,
		ValidateFunc: validation.IntAtLeast(0),
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
}

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

		Schema: mergeSchema(replicationSchemaCommon,repMultipleSchema),
	}
}

func unpackReplicationConfig(s *schema.ResourceData) ReplicationConfig {
	d := &ResourceData{s}
	replicationConfig := new(ReplicationConfig)

	repo := d.getString("repo_key", false)

	if v, ok := d.GetOkExists("replications"); ok {
		arr := v.([]interface{})

		tmp := make([]utils.ReplicationBody, 0, len(arr))
		replicationConfig.Replications = tmp

		for i, o := range arr {
			if i == 0 {
				replicationConfig.RepoKey = repo
				replicationConfig.CronExp = d.getString("cron_exp", false)
				replicationConfig.EnableEventReplication = d.getBool("enable_event_replication", false)
			}

			m := o.(map[string]interface{})

			var replication utils.ReplicationBody

			replication.RepoKey = repo

			if v, ok = m["url"]; ok {
				replication.URL = v.(string)
			}

			if v, ok = m["socket_timeout_millis"]; ok {
				replication.SocketTimeoutMillis = v.(int)
			}

			if v, ok = m["username"]; ok {
				replication.Username = v.(string)
			}

			if v, ok = m["enabled"]; ok {
				replication.Enabled = v.(bool)
			}

			if v, ok = m["sync_deletes"]; ok {
				replication.SyncDeletes = v.(bool)
			}

			if v, ok = m["sync_properties"]; ok {
				replication.SyncProperties = v.(bool)
			}

			if v, ok = m["sync_statistics"]; ok {
				replication.SyncStatistics = v.(bool)
			}

			if prefix, ok := m["path_prefix"]; ok {
				replication.PathPrefix = prefix.(string)
			}

			if pass, ok := m["password"]; ok {
				replication.Password = pass.(string)
			}

			replicationConfig.Replications = append(replicationConfig.Replications, replication)
		}
	}

	return *replicationConfig
}

func packReplicationConfig(replicationConfig *ReplicationConfig, d *schema.ResourceData) error {
	var errors []error
	setValue := mkLens(d)

	setValue("repo_key", replicationConfig.RepoKey)
	setValue("cron_exp", replicationConfig.CronExp)
	errors = setValue("enable_event_replication", replicationConfig.EnableEventReplication)

	if replicationConfig.Replications != nil {
		var replications []map[string]interface{}
		for _, repo := range replicationConfig.Replications {
			replication := make(map[string]interface{})

			replication["url"] = repo.URL
			replication["socket_timeout_millis"] = repo.SocketTimeoutMillis
			replication["username"] = repo.Username
			replication["password"] = getMD5Hash(repo.Password)
			replication["enabled"] = repo.Enabled
			replication["sync_deletes"] = repo.SyncDeletes
			replication["sync_properties"] = repo.SyncProperties
			replication["sync_statistics"] = repo.SyncStatistics
			replication["path_prefix"] = repo.PathPrefix
			replications = append(replications, replication)
		}

		errors = setValue("replications", replications)
	}

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourceReplicationConfigCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	replicationConfig := unpackReplicationConfig(d)

	_, err := client.R().SetBody(replicationConfig).Put("artifactory/api/replications/multiple/" + replicationConfig.RepoKey)
	if err != nil {
		return err
	}

	d.SetId(replicationConfig.RepoKey)
	return resourceReplicationConfigRead(d, m)
}

func resourceReplicationConfigRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Resty
	var replications []utils.ReplicationBody
	_, err := c.R().SetResult(&replications).Get("artifactory/api/replications/" + d.Id())

	if err != nil {
		return err
	}

	repConfig := ReplicationConfig{
		RepoKey:      d.Id(),
		Replications: replications,
	}
	if len(replications) > 0{
		repConfig.EnableEventReplication = replications[0].EnableEventReplication
		repConfig.CronExp  = replications[0].CronExp
	}
	return packReplicationConfig(&repConfig, d)
}

func resourceReplicationConfigUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	replicationConfig := unpackReplicationConfig(d)

	_, err := client.R().SetBody(replicationConfig).Post("/api/replications/" + d.Id())
	if err != nil {
		return err
	}

	d.SetId(replicationConfig.RepoKey)

	return resourceReplicationConfigRead(d, m)
}

func resourceReplicationConfigDelete(d *schema.ResourceData, m interface{}) error {
	_, err := m.(*ArtClient).Resty.R().Delete("artifactory/api/replications/" + d.Id())
	return err
}
func repConfigExists(id string, m interface{}) (bool, error) {
	_, err := m.(*ArtClient).Resty.R().Head("artifactory/api/replications/" + id)
	return err == nil, err
}

func resourceReplicationConfigExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return repConfigExists(d.Id(),m)
}
