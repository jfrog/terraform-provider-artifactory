package replication

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
	"golang.org/x/exp/slices"
)

type ReplicationBody struct {
	Username                        string `json:"username"`
	Password                        string `json:"password"`
	URL                             string `json:"url"`
	CronExp                         string `json:"cronExp"`
	RepoKey                         string `json:"repoKey"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	SocketTimeoutMillis             int    `json:"socketTimeoutMillis"`
	Enabled                         bool   `json:"enabled"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	SyncStatistics                  bool   `json:"syncStatistics"`
	PathPrefix                      string `json:"pathPrefix"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
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

var pushReplicationSchema = map[string]*schema.Schema{
	"url": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
	},
	"socket_timeout_millis": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
	},
	"username": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Username for push replication",
	},
	"password": {
		Type:             schema.TypeString,
		Required:         true,
		Sensitive:        true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Password for push replication",
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
	"check_binary_existence_in_filestore": {
		Type:     schema.TypeBool,
		Optional: true,
		Description: "When true, enables distributed checksum storage. For more information, see " +
			"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
	},
}

var pushRepMultipleSchema = map[string]*schema.Schema{
	"cron_exp": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validator.CronLength,
		Description:      "Cron expression to control the operation frequency.",
	},
	"repo_key": {
		Type:     schema.TypeString,
		Required: true,
	},
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
	},
	"replications": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: pushReplicationSchema,
		},
	},
}

func unpackPushReplication(s *schema.ResourceData) UpdatePushReplication {
	d := &util.ResourceData{ResourceData: s}
	pushReplication := new(UpdatePushReplication)

	repo := d.GetString("repo_key", false)

	if v, ok := d.GetOk("replications"); ok {
		arr := v.([]interface{})

		tmp := make([]updateReplicationBody, 0, len(arr))
		pushReplication.Replications = tmp

		for i, o := range arr {
			if i == 0 {
				pushReplication.RepoKey = repo
				pushReplication.CronExp = d.GetString("cron_exp", false)
				pushReplication.EnableEventReplication = d.GetBool("enable_event_replication", false)
			}

			m := o.(map[string]interface{})

			var replication updateReplicationBody

			replication.RepoKey = repo
			replication.CronExp = d.GetString("cron_exp", false)

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

			if v, ok = m["check_binary_existence_in_filestore"]; ok {
				replication.CheckBinaryExistenceInFilestore = v.(bool)
			}

			pushReplication.Replications = append(pushReplication.Replications, replication)
		}
	}

	return *pushReplication
}

func packPushReplication(pushReplication *GetPushReplication, d *schema.ResourceData) diag.Diagnostics {
	var errors []error
	setValue := util.MkLens(d)

	setValue("repo_key", pushReplication.RepoKey)
	setValue("cron_exp", pushReplication.CronExp)
	errors = setValue("enable_event_replication", pushReplication.EnableEventReplication)

	if pushReplication.Replications != nil {

		// Get replications from TF state
		var tfReplications []interface{}
		if v, ok := d.GetOkExists("replications"); ok {
			tfReplications = v.([]interface{})
		}

		var replications []map[string]interface{}
		for _, repl := range pushReplication.Replications {
			replication := make(map[string]interface{})

			replication["url"] = repl.URL
			replication["socket_timeout_millis"] = repl.SocketTimeoutMillis
			replication["username"] = repl.Username

			// find the matching replication from current state
			tfReplicationIndex := slices.IndexFunc(tfReplications, func(r interface{}) bool {
				return r.(map[string]interface{})["url"] == repl.URL
			})
			if tfReplicationIndex != -1 {
				// set password from current state to avoid state drift
				// from missing password in Artifactory API response
				replication["password"] = tfReplications[tfReplicationIndex].(map[string]interface{})["password"]
			}

			replication["enabled"] = repl.Enabled
			replication["sync_deletes"] = repl.SyncDeletes
			replication["sync_properties"] = repl.SyncProperties
			replication["sync_statistics"] = repl.SyncStatistics
			replication["path_prefix"] = repl.PathPrefix
			replication["proxy"] = repl.ProxyRef
			replication["check_binary_existence_in_filestore"] = repl.CheckBinaryExistenceInFilestore
			replications = append(replications, replication)
		}

		errors = setValue("replications", replications)
	}
	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourcePushReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackPushReplication(d)

	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		Put(EndpointPath + "multiple/" + pushReplication.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pushReplication.RepoKey)
	return resourcePushReplicationRead(ctx, d, m)
}

func resourcePushReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(util.ProvderMetadata).Client
	var replications []getReplicationBody
	_, err := c.R().SetResult(&replications).Get(EndpointPath + d.Id())

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

	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		AddRetryCondition(client.RetryOnMergeError).
		Post(EndpointPath + "multiple/" + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePushReplicationRead(ctx, d, m)
}

func ResourceArtifactoryPushReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePushReplicationCreate,
		ReadContext:   resourcePushReplicationRead,
		UpdateContext: resourcePushReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:             pushRepMultipleSchema,
		Description:        "Add or replace multiple replication configurations for given repository key. Supported by local repositories. Artifactory Enterprise license is required.",
		DeprecationMessage: "This resource is replaced by `artifactory_local_repository_multi_replication` for clarity. All the attributes are identical, please consider the migration.",
	}
}
