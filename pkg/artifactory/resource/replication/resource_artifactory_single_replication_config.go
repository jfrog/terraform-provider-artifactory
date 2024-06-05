package replication

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func ResourceArtifactorySingleReplicationConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceSingleReplicationConfigCreate,
		ReadContext:   resourceSingleReplicationConfigRead,
		UpdateContext: resourceSingleReplicationConfigUpdate,
		DeleteContext: resourceReplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: utilsdk.MergeMaps(replicationSchemaCommon, replicationSchema),
		Description: "Used for configuring replications on repos. However, the TCL only makes " +
			"good sense for local repo replication (PUSH) and not remote (PULL).",
		DeprecationMessage: "This resource has been deprecated in favor of the more explicitly name" +
			"artifactory_pull_replication resource.",
	}
}

func resourceSingleReplicationConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_single_replication_config deprecated. Use artifactory_pull_replication instead")
}

func resourceSingleReplicationConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_single_replication_config deprecated. Use artifactory_pull_replication instead")
}

func resourceSingleReplicationConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_single_replication_config deprecated. Use artifactory_pull_replication instead")
}
