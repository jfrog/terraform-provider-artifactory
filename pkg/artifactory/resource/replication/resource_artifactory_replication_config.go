package replication

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"

	"github.com/jfrog/terraform-provider-shared/validator"
)

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
		DeprecationMessage: "This resource has been deprecated in favor of the more explicitly name" +
			"artifactory_push_replication resource.",
	}
}

func resourceReplicationConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_replication_config deprecated. Use artifactory_push_replication instead")
}

func resourceReplicationConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_replication_config deprecated. Use artifactory_push_replication instead")
}

func resourceReplicationConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_replication_config deprecated. Use artifactory_push_replication instead")
}
