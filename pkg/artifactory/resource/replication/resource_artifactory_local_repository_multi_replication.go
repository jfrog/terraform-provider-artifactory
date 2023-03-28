package replication

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
	"golang.org/x/exp/slices"
)

type localMultiReplicationBody struct {
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
	IncludePathPrefixPattern        string `json:"includePathPrefixPattern"`
	ExcludePathPrefixPattern        string `json:"excludePathPrefixPattern"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
}

type getLocalMultiReplicationBody struct {
	localMultiReplicationBody
	ProxyRef       string `json:"proxyRef"`
	ReplicationKey string `json:"replicationKey"`
}

type updateLocalMultiReplicationBody struct {
	localMultiReplicationBody
	Proxy string `json:"proxy"`
}

type GetLocalMultiReplication struct {
	RepoKey                string                         `json:"-"`
	CronExp                string                         `json:"cronExp,omitempty"`
	EnableEventReplication bool                           `json:"enableEventReplication"`
	Replications           []getLocalMultiReplicationBody `json:"replications,omitempty"`
}

type UpdateLocalMultiReplication struct {
	RepoKey                string                            `json:"-"`
	CronExp                string                            `json:"cronExp,omitempty"`
	EnableEventReplication bool                              `json:"enableEventReplication"`
	Replications           []updateLocalMultiReplicationBody `json:"replications,omitempty"`
}

var localReplicationSchema = map[string]*schema.Schema{
	"url": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
		Description:      "The URL of the target local repository on a remote Artifactory server. Use the format `https://<artifactory_url>/artifactory/<repository_name>`.",
	},
	"socket_timeout_millis": {
		Type:             schema.TypeInt,
		Optional:         true,
		Default:          15000,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
		Description:      "The network timeout in milliseconds to use for remote operations.",
	},
	"username": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "The username on the remote Artifactory instance.",
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Use either the HTTP authentication password or identity token (https://www.jfrog.com/confluence/display/JFROG/User+Profile#UserProfile-IdentityTokenidentitytoken).",
	},
	"sync_deletes": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, items that were deleted locally should also be deleted remotely (also applies to properties metadata). Note that enabling this option, will delete artifacts on the target that do not exist in the source repository. Default value is `false`.",
	},
	"sync_properties": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "When set, the task also synchronizes the properties of replicated artifacts. Default value is `true`.",
	},
	"sync_statistics": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, the task also synchronizes artifact download statistics. Set to avoid inadvertent cleanup at the target instance when setting up replication for disaster recovery. Default value is `false`.",
	},
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "When set, enables replication of this repository to the target specified in `url` attribute. Default value is `true`.",
	},
	"include_path_prefix_pattern": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"exclude_path_prefix_pattern": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default, no artifacts are excluded.",
	},
	"proxy": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Proxy key from Artifactory Proxies settings. The proxy configuration will be used when communicating with the remote instance.",
	},
	"replication_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Replication ID. The ID is known only after the replication is created, for this reason it's `Computed` and can not be set by the user in HCL.",
	},
	"check_binary_existence_in_filestore": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enabling the `check_binary_existence_in_filestore` flag requires an Enterprise Plus license. When true, enables distributed checksum storage. For more information, see [Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
	},
}

var localMultiReplicationSchema = map[string]*schema.Schema{
	"repo_key": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Repository name.",
	},
	"cron_exp": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validator.CronLength,
		Description:      "Cron expression to control the operation frequency.",
	},
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
	},
	"replication": {
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &schema.Resource{
			Schema: localReplicationSchema,
		},
	},
}

func unpackLocalMultiReplication(s *schema.ResourceData) UpdateLocalMultiReplication {
	d := &util.ResourceData{ResourceData: s}
	pushReplication := new(UpdateLocalMultiReplication)

	repo := d.GetString("repo_key", false)

	if v, ok := d.GetOk("replication"); ok {
		arr := v.([]interface{})

		tmp := make([]updateLocalMultiReplicationBody, 0, len(arr))
		pushReplication.Replications = tmp

		for i, o := range arr {
			if i == 0 {
				pushReplication.RepoKey = repo
				pushReplication.CronExp = d.GetString("cron_exp", false)
				pushReplication.EnableEventReplication = d.GetBool("enable_event_replication", false)
			}

			m := o.(map[string]interface{})

			var replication updateLocalMultiReplicationBody

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

			if include, ok := m["include_path_prefix_pattern"]; ok {
				replication.IncludePathPrefixPattern = include.(string)
			}

			if exclude, ok := m["exclude_path_prefix_pattern"]; ok {
				replication.ExcludePathPrefixPattern = exclude.(string)
			}

			if _, ok := m["proxy"]; ok {
				replication.Proxy = repository.HandleResetWithNonExistentValue(d, fmt.Sprintf("replication.%d.proxy", i))
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

func packLocalMultiReplication(pushReplication *GetLocalMultiReplication, d *schema.ResourceData) diag.Diagnostics {
	var errors []error
	setValue := util.MkLens(d)

	setValue("repo_key", pushReplication.RepoKey)
	setValue("cron_exp", pushReplication.CronExp)
	errors = setValue("enable_event_replication", pushReplication.EnableEventReplication)

	if pushReplication.Replications != nil {

		// Get replication list from TF state
		var tfReplications []interface{}
		if v, ok := d.GetOkExists("replication"); ok {
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
			replication["include_path_prefix_pattern"] = repl.IncludePathPrefixPattern
			replication["exclude_path_prefix_pattern"] = repl.ExcludePathPrefixPattern
			replication["proxy"] = repl.ProxyRef
			replication["replication_key"] = repl.ReplicationKey
			replication["check_binary_existence_in_filestore"] = repl.CheckBinaryExistenceInFilestore
			replications = append(replications, replication)
		}

		errors = setValue("replication", replications)
	}
	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourceLocalMultiReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackLocalMultiReplication(d)

	if verified, err := verifyRepoRclass(pushReplication.RepoKey, "local", m); !verified {
		return diag.Errorf("source repository rclass is not local, only remote repositories are supported by this resource %v", err)
	}
	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		Put(EndpointPath + "multiple/" + pushReplication.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pushReplication.RepoKey)
	return resourceLocalMultiReplicationRead(ctx, d, m)
}

func resourceLocalMultiReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(util.ProvderMetadata).Client
	var replications []getLocalMultiReplicationBody
	resp, err := c.R().SetResult(&replications).Get(EndpointPath + d.Id())

	if err != nil {
		if resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	repConfig := GetLocalMultiReplication{
		RepoKey:      d.Id(),
		Replications: replications,
	}
	if len(replications) > 0 {
		repConfig.EnableEventReplication = replications[0].EnableEventReplication
		repConfig.CronExp = replications[0].CronExp
	}
	return packLocalMultiReplication(&repConfig, d)
}

func resourceLocalMultiReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackLocalMultiReplication(d)

	if verified, err := verifyRepoRclass(pushReplication.RepoKey, "local", m); !verified {
		return diag.Errorf("source repository rclass is not local, only remote repositories are supported by this resource %v", err)
	}
	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		AddRetryCondition(client.RetryOnMergeError).
		Post(EndpointPath + "multiple/" + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceLocalMultiReplicationRead(ctx, d, m)
}

func ResourceArtifactoryLocalRepositoryMultiReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLocalMultiReplicationCreate,
		ReadContext:   resourceLocalMultiReplicationRead,
		UpdateContext: resourceLocalMultiReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "Add or replace multiple replication configurations for given repository key. Supported by local repositories. Artifactory Enterprise license is required.",
		Schema:      localMultiReplicationSchema,
	}
}
