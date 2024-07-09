package webhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type ReleaseBundleV2PromotionWebhookCriteria struct {
	SelectedEnvironments []string `json:"selectedEnvironments"`
}

var releaseBundleV2PromotionWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"selected_environments": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of environments",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied, on which release bundles promotion.",
		},
	})
}

var packReleaseBundleV2PromotionCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"selected_environments": schema.NewSet(schema.HashString, artifactoryCriteria["selectedEnvironments"].([]interface{})),
	}
}

var unpackReleaseBundleV2PromotionCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseWebhookCriteria) interface{} {
	return ReleaseBundleV2PromotionWebhookCriteria{
		SelectedEnvironments: utilsdk.CastToStringArr(terraformCriteria["selected_environments"].(*schema.Set).List()),
	}
}

var releaseBundleV2PromotionCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	return nil
}
