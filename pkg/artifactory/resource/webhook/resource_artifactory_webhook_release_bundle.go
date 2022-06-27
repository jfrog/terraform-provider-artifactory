package webhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
)

type ReleaseBundleWebhookCriteria struct {
	BaseWebhookCriteria
	AnyReleaseBundle              bool     `json:"anyReleaseBundle"`
	RegisteredReleaseBundlesNames []string `json:"registeredReleaseBundlesNames"`
}

var releaseBundleWebhookSchema = func(webhookType string, version int) map[string]*schema.Schema {
	return util.MergeMaps(getBaseSchemaByVersion(webhookType, version), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: util.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
					"any_release_bundle": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any release bundles or distributions",
					},
					"registered_release_bundle_names": {
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

var packReleaseBundleCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"any_release_bundle":              artifactoryCriteria["anyReleaseBundle"].(bool),
		"registered_release_bundle_names": schema.NewSet(schema.HashString, artifactoryCriteria["registeredReleaseBundlesNames"].([]interface{})),
	}
}

var unpackReleaseBundleCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseWebhookCriteria) interface{} {
	return ReleaseBundleWebhookCriteria{
		AnyReleaseBundle:              terraformCriteria["any_release_bundle"].(bool),
		RegisteredReleaseBundlesNames: util.CastToStringArr(terraformCriteria["registered_release_bundle_names"].(*schema.Set).List()),
		BaseWebhookCriteria:           baseCriteria,
	}
}

var releaseBundleCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	tflog.Debug(ctx, "releaseBundleCriteriaValidation")

	anyReleaseBundle := criteria["any_release_bundle"].(bool)
	registeredReleaseBundlesNames := criteria["registered_release_bundle_names"].(*schema.Set).List()

	if anyReleaseBundle == false && len(registeredReleaseBundlesNames) == 0 {
		return fmt.Errorf("registered_release_bundle_names cannot be empty when any_release_bundle is false")
	}

	return nil
}
