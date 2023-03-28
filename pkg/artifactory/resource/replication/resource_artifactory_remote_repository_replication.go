package replication

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
)

type remoteReplicationBody struct {
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	IncludePathPrefixPattern        string `json:"includePathPrefixPattern"`
	ExcludePathPrefixPattern        string `json:"excludePathPrefixPattern"`
	RepoKey                         string `json:"repoKey"`
	ReplicationKey                  string `json:"replicationKey"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
}

type getRemoteReplicationBody struct {
	remoteReplicationBody
	ReplicationKey string `json:"replicationKey"`
}

type updateRemoteReplicationBody struct {
	remoteReplicationBody
}

var remoteReplicationSchema = map[string]*schema.Schema{
	"repo_key": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Repository name.",
	},
	"cron_exp": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "The Cron expression that determines when the next replication will be triggered.",
	},
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
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
		Description: "When set, the task also synchronizes the properties of replicated artifacts. Default value is `true`",
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
		Description:      "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*. By default no artifacts are excluded.",
	},
	"replication_key": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: "Replication ID.",
	},
	"check_binary_existence_in_filestore": {
		Type:     schema.TypeBool,
		Optional: true,
		Description: "Enabling the `check_binary_existence_in_filestore` flag requires an Enterprise Plus license. When true, enables distributed checksum storage. For more information, see " +
			"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
	},
}

func unpackRemoteReplication(s *schema.ResourceData) updateRemoteReplicationBody {
	d := &util.ResourceData{ResourceData: s}

	return updateRemoteReplicationBody{
		remoteReplicationBody: remoteReplicationBody{
			EnableEventReplication:          d.GetBool("enable_event_replication", false),
			Enabled:                         d.GetBool("enabled", false),
			CronExp:                         d.GetString("cron_exp", false),
			SyncDeletes:                     d.GetBool("sync_deletes", false),
			SyncProperties:                  d.GetBool("sync_properties", false),
			RepoKey:                         d.GetString("repo_key", false),
			IncludePathPrefixPattern:        d.GetString("include_path_prefix_pattern", false),
			ExcludePathPrefixPattern:        d.GetString("exclude_path_prefix_pattern", false),
			CheckBinaryExistenceInFilestore: d.GetBool("check_binary_existence_in_filestore", false),
		},
	}
}

func packRemoteReplication(remoteReplication *getRemoteReplicationBody, d *schema.ResourceData) diag.Diagnostics {

	var errors []error
	setValue := util.MkLens(d)
	setValue("enable_event_replication", remoteReplication.EnableEventReplication)
	setValue("enabled", remoteReplication.Enabled)
	setValue("cron_exp", remoteReplication.CronExp)
	setValue("sync_deletes", remoteReplication.SyncDeletes)
	setValue("sync_properties", remoteReplication.SyncProperties)
	setValue("repo_key", remoteReplication.RepoKey)
	setValue("include_path_prefix_pattern", remoteReplication.IncludePathPrefixPattern)
	setValue("exclude_path_prefix_pattern", remoteReplication.ExcludePathPrefixPattern)
	setValue("replication_key", remoteReplication.ReplicationKey)
	errors = setValue("check_binary_existence_in_filestore", remoteReplication.CheckBinaryExistenceInFilestore)

	if errors != nil && len(errors) > 0 {
		return diag.Errorf("failed to pack replication config %q", errors)
	}

	return nil
}

func resourceRemoteReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackRemoteReplication(d)

	if verified, err := verifyRepoRclass(pushReplication.RepoKey, "remote", m); !verified {
		return diag.Errorf("source repository rclass is not remote or can't be verified, only remote repositories are supported by this resource: %v", err)
	}
	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		Put(EndpointPath + pushReplication.RepoKey)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(pushReplication.RepoKey)
	return resourceRemoteReplicationRead(ctx, d, m)
}

func resourceRemoteReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(util.ProvderMetadata).Client

	var replication getRemoteReplicationBody

	resp, err := c.R().SetResult(&replication).Get(EndpointPath + d.Id())

	if err != nil {
		if resp != nil && (resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return packRemoteReplication(&replication, d)
}

func resourceRemoteReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pushReplication := unpackRemoteReplication(d)

	if verified, err := verifyRepoRclass(pushReplication.RepoKey, "remote", m); !verified {
		return diag.Errorf("source repository rclass is not remote or can't be verified, only remote repositories are supported by this resource: %v", err)
	}
	_, err := m.(util.ProvderMetadata).Client.R().
		SetBody(pushReplication).
		AddRetryCondition(client.RetryOnMergeError).
		Post(EndpointPath + d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRemoteReplicationRead(ctx, d, m)
}

func ResourceArtifactoryRemoteRepositoryReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRemoteReplicationCreate,
		ReadContext:   resourceRemoteReplicationRead,
		UpdateContext: resourceRemoteReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Description: "Add or replace a single replication configuration for given repository key. Supported by remote repositories. Artifactory Pro license is required.",
		Schema:      remoteReplicationSchema,
	}
}
