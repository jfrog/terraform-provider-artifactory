package artifactory

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ReplicationBody struct {
	Username               string `json:"username"`
	Password               string `json:"password"`
	URL                    string `json:"url"`
	CronExp                string `json:"cronExp"`
	RepoKey                string `json:"repoKey"`
	EnableEventReplication bool   `json:"enableEventReplication"`
	SocketTimeoutMillis    int    `json:"socketTimeoutMillis"`
	Enabled                bool   `json:"enabled"`
	SyncDeletes            bool   `json:"syncDeletes"`
	SyncProperties         bool   `json:"syncProperties"`
	SyncStatistics         bool   `json:"syncStatistics"`
	PathPrefix             string `json:"pathPrefix"`
}

type getReplicationBody struct {
	ReplicationBody
	ProxyRef string `json:"proxyRef"`
}

type updateReplicationBody struct {
	ReplicationBody
	Proxy string `json:"proxy"`
}

type GetPushReplication struct {
	RepoKey                string               `json:"-"`
	CronExp                string               `json:"cronExp,omitempty"`
	EnableEventReplication bool                 `json:"enableEventReplication,omitempty"`
	Replications           []getReplicationBody `json:"replications,omitempty"`
}

type UpdatePushReplication struct {
	RepoKey                string                  `json:"-"`
	CronExp                string                  `json:"cronExp,omitempty"`
	EnableEventReplication bool                    `json:"enableEventReplication,omitempty"`
	Replications           []updateReplicationBody `json:"replications,omitempty"`
}

var pushReplicationSchemaCommon = map[string]*schema.Schema{
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

var pushRepMultipleSchema = map[string]*schema.Schema{
	"replications": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: pushReplicationSchema,
		},
	},
}

var pushReplicationSchema = map[string]*schema.Schema{
	"url": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
	},
	"socket_timeout_millis": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validation.IntAtLeast(0),
	},
	"username": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"password": {
		Type:      schema.TypeString,
		Computed:  true,
		Sensitive: true,
		Description: "If a password is used to create the resource, it will be returned as encrypted and this will become the new state." +
			"Practically speaking, what this means is that, the password can only be set, not gotten. ",
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

func resourceArtifactoryPushReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePushReplicationCreate,
		ReadContext:   resourcePushReplicationRead,
		UpdateContext: resourcePushReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: mergeSchema(pushReplicationSchemaCommon, pushRepMultipleSchema),
	}
}

func unpackPushReplication(s *schema.ResourceData) UpdatePushReplication {
	d := &ResourceData{s}
	pushReplication := new(UpdatePushReplication)

	repo := d.getString("repo_key", false)

	if v, ok := d.GetOk("replications"); ok {
		arr := v.([]interface{})

		tmp := make([]updateReplicationBody, 0, len(arr))
		pushReplication.Replications = tmp

		for i, o := range arr {
			if i == 0 {
				pushReplication.RepoKey = repo
				pushReplication.CronExp = d.getString("cron_exp", false)
				pushReplication.EnableEventReplication = d.getBool("enable_event_replication", false)
			}

			m := o.(map[string]interface{})

			var replication updateReplicationBody

			replication.RepoKey = repo
			replication.CronExp = d.getString("cron_exp", false)

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

			if _, ok := m["proxy"]; ok {
				replication.Proxy = handleResetWithNonExistantValue(d, fmt.Sprintf("replications.%d.proxy", i))
			}

			if pass, ok := m["password"]; ok {
				replication.Password = pass.(string)
			}

			pushReplication.Replications = append(pushReplication.Replications, replication)
		}
	}

	return *pushReplication
}

func packPushReplication(pushReplication *GetPushReplication, d *schema.ResourceData) diag.Diagnostics {
	var errors []error
	setValue := mkLens(d)

	setValue("repo_key", pushReplication.RepoKey)
	setValue("cron_exp", pushReplication.CronExp)
	errors = setValue("enable_event_replication", pushReplication.EnableEventReplication)

	if pushReplication.Replications != nil {
		var replications []map[string]interface{}
		for _, repo := range pushReplication.Replications {
			replication := make(map[string]interface{})

			replication["url"] = repo.URL
			replication["socket_timeout_millis"] = repo.SocketTimeoutMillis
			replication["username"] = repo.Username
			replication["password"] = repo.Password
			replication["enabled"] = repo.Enabled
			replication["sync_deletes"] = repo.SyncDeletes
			replication["sync_properties"] = repo.SyncProperties
			replication["sync_statistics"] = repo.SyncStatistics
			replication["path_prefix"] = repo.PathPrefix
			replication["proxy"] = repo.ProxyRef
			replications = append(replications, replication)
		}

		errors = setValue("replications", replications)
	}
	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

const apiPath = "artifactory/api/replications/"

func resourcePushReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackPushReplication(d)

	_, err := m.(*resty.Client).R().SetBody(pushReplication).Put(apiPath + "multiple/" + pushReplication.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pushReplication.RepoKey)
	return resourcePushReplicationRead(ctx, d, m)
}

func resourcePushReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*resty.Client)
	var replications []getReplicationBody
	_, err := c.R().SetResult(&replications).Get(apiPath + d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	repConfig := GetPushReplication{
		RepoKey:      d.Id(),
		Replications: replications,
	}
	if len(replications) > 0 {
		repConfig.EnableEventReplication = replications[0].EnableEventReplication
		repConfig.CronExp = replications[0].CronExp
	}
	return packPushReplication(&repConfig, d)
}

func resourcePushReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackPushReplication(d)

	_, err := m.(*resty.Client).R().SetBody(pushReplication).Post(apiPath + "multiple/" + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// d.SetId(pushReplication.RepoKey)

	return resourcePushReplicationRead(ctx, d, m)
}

func resourceReplicationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	_, err := m.(*resty.Client).R().Delete(apiPath + d.Id())
	return diag.FromErr(err)
}

func repConfigExists(id string, m interface{}) (bool, error) {
	_, err := m.(*resty.Client).R().Head(apiPath + id)
	return err == nil, err
}
