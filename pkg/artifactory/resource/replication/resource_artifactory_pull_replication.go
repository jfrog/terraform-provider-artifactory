package replication

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourcePullReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_pull_replication deprecated. Use artifactory_remote_repository_replication instead")
}

func resourcePullReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_pull_replication deprecated. Use artifactory_remote_repository_replication instead")
}

func resourcePullReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_pull_replication deprecated. Use artifactory_remote_repository_replication instead")
}

var pullReplicationSchema = map[string]*schema.Schema{
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
	"url": {
		Type:             schema.TypeString,
		Optional:         true,
		ForceNew:         true,
		RequiredWith:     []string{"username", "password"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
		Description:      "URL for local repository replication. Required for local repository, but not needed for remote repository.",
	},
	"socket_timeout_millis": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
	},
	"username": {
		Type:             schema.TypeString,
		Optional:         true,
		RequiredWith:     []string{"url", "password"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Username for local repository replication. Required for local repository, but not needed for remote repository.",
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		RequiredWith:     []string{"url", "username"},
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Password for local repository replication. Required for local repository, but not needed for remote repository.",
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
		Default:  false,
		Description: "When true, enables distributed checksum storage. For more information, see " +
			"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
	},
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
	},
}

func ResourceArtifactoryPullReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePullReplicationCreate,
		ReadContext:   resourcePullReplicationRead,
		UpdateContext: resourcePullReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Schema:             pullReplicationSchema,
		Description:        "Used for configuring pull replication on local or remote repos.",
		DeprecationMessage: "This resource is replaced by `artifactory_remote_repository_replication` for clarity. Please consider the migration.",
	}
}
