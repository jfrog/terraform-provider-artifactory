package replication

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/validator"
)

type GetReplicationConfig struct {
	RepoKey                string               `json:"-"`
	CronExp                string               `json:"cronExp,omitempty"`
	EnableEventReplication bool                 `json:"enableEventReplication,omitempty"`
	Replications           []getReplicationBody `json:"replications,omitempty"`
}

type UpdateReplicationConfig struct {
	RepoKey                string                  `json:"-"`
	CronExp                string                  `json:"cronExp,omitempty"`
	EnableEventReplication bool                    `json:"enableEventReplication,omitempty"`
	Replications           []updateReplicationBody `json:"replications,omitempty"`
}

var replicationSchemaCommon = map[string]*schema.Schema{
	"repo_key": {
		Type:     schema.TypeString,
		Required: true,
	},
	"cron_exp": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validator.CronLength,
		Description:      "Cron expression to control the operation frequency.",
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
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Proxy key from Artifactory Proxies setting",
	},
}

func ResourceArtifactoryReplicationConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceReplicationConfigCreate,
		ReadContext:   resourceReplicationConfigRead,
		UpdateContext: resourceReplicationConfigUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: utilsdk.MergeMaps(replicationSchemaCommon, repMultipleSchema),
		DeprecationMessage: "This resource has been deprecated in favour of the more explicitly name" +
			"artifactory_push_replication resource.",
	}
}

func unpackReplicationConfig(s *schema.ResourceData) UpdateReplicationConfig {
	d := &utilsdk.ResourceData{ResourceData: s}
	replicationConfig := new(UpdateReplicationConfig)

	repo := d.GetString("repo_key", false)

	if v, ok := d.GetOk("replications"); ok {
		arr := v.([]interface{})

		tmp := make([]updateReplicationBody, 0, len(arr))
		replicationConfig.Replications = tmp

		for i, o := range arr {
			if i == 0 {
				replicationConfig.RepoKey = repo
				replicationConfig.CronExp = d.GetString("cron_exp", false)
				replicationConfig.EnableEventReplication = d.GetBool("enable_event_replication", false)
			}

			m := o.(map[string]interface{})

			var replication updateReplicationBody

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

			if _, ok := m["proxy"]; ok {
				replication.Proxy = repository.HandleResetWithNonExistentValue(d, fmt.Sprintf("replications.%d.proxy", i))
			}

			if pass, ok := m["password"]; ok {
				replication.Password = pass.(string)
			}

			replicationConfig.Replications = append(replicationConfig.Replications, replication)
		}
	}

	return *replicationConfig
}

func packReplicationConfig(replicationConfig *GetReplicationConfig, d *schema.ResourceData) diag.Diagnostics {
	var errors []error
	setValue := utilsdk.MkLens(d)

	setValue("repo_key", replicationConfig.RepoKey)
	setValue("cron_exp", replicationConfig.CronExp)
	errors = setValue("enable_event_replication", replicationConfig.EnableEventReplication)

	if replicationConfig.Replications != nil {
		var replications []map[string]interface{}
		for _, repl := range replicationConfig.Replications {
			replication := make(map[string]interface{})

			replication["url"] = repl.URL
			replication["socket_timeout_millis"] = repl.SocketTimeoutMillis
			replication["username"] = repl.Username
			replication["enabled"] = repl.Enabled
			replication["sync_deletes"] = repl.SyncDeletes
			replication["sync_properties"] = repl.SyncProperties
			replication["sync_statistics"] = repl.SyncStatistics
			replication["path_prefix"] = repl.PathPrefix
			replication["proxy"] = repl.ProxyRef

			replications = append(replications, replication)
		}

		errors = setValue("replications", replications)
	}
	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourceReplicationConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackReplicationConfig(d)

	_, err := m.(utilsdk.ProvderMetadata).Client.R().
		SetBody(replicationConfig).
		Put(EndpointPath + "multiple/" + replicationConfig.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)
	return resourceReplicationConfigRead(ctx, d, m)
}

func resourceReplicationConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(utilsdk.ProvderMetadata).Client
	var replications []getReplicationBody
	_, err := c.R().SetResult(&replications).Get(EndpointPath + d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	repConfig := GetReplicationConfig{
		RepoKey:      d.Id(),
		Replications: replications,
	}
	if len(replications) > 0 {
		repConfig.EnableEventReplication = replications[0].EnableEventReplication
		repConfig.CronExp = replications[0].CronExp
	}
	return packReplicationConfig(&repConfig, d)
}

func resourceReplicationConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	replicationConfig := unpackReplicationConfig(d)

	_, err := m.(utilsdk.ProvderMetadata).Client.R().SetBody(replicationConfig).Post(EndpointPath + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(replicationConfig.RepoKey)

	return resourceReplicationConfigRead(ctx, d, m)
}
