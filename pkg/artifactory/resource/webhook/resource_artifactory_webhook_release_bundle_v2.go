package webhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type ReleaseBundleV2WebhookCriteria struct {
	BaseCriteriaAPIModel
	AnyReleaseBundle       bool     `json:"anyReleaseBundle"`
	SelectedReleaseBundles []string `json:"selectedReleaseBundles"`
}

var releaseBundleV2WebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"any_release_bundle": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any release bundles or distributions",
					},
					"selected_release_bundles": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of release bundle names",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied, on which release bundles or distributions.",
		},
	})
}

var packReleaseBundleV2Criteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"any_release_bundle":       artifactoryCriteria["anyReleaseBundle"].(bool),
		"selected_release_bundles": schema.NewSet(schema.HashString, artifactoryCriteria["selectedReleaseBundles"].([]interface{})),
	}
}

var unpackReleaseBundleV2Criteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return ReleaseBundleV2WebhookCriteria{
		AnyReleaseBundle:       terraformCriteria["any_release_bundle"].(bool),
		SelectedReleaseBundles: utilsdk.CastToStringArr(terraformCriteria["selected_release_bundles"].(*schema.Set).List()),
		BaseCriteriaAPIModel:   baseCriteria,
	}
}

var releaseBundleV2CriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	anyReleaseBundle := criteria["any_release_bundle"].(bool)
	selectedReleaseBundles := criteria["selected_release_bundles"].(*schema.Set).List()

	if !anyReleaseBundle && len(selectedReleaseBundles) == 0 {
		return fmt.Errorf("selected_release_bundles cannot be empty when any_release_bundle is false")
	}

	return nil
}
