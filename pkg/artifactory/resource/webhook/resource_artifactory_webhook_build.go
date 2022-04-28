package webhook

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
)

type BuildWebhookCriteria struct {
	BaseWebhookCriteria
	AnyBuild       bool     `json:"anyBuild"`
	SelectedBuilds []string `json:"selectedBuilds"`
}

var buildWebhookSchema = func(webhookType string) map[string]*schema.Schema {
	return util.MergeSchema(baseWebhookBaseSchema(webhookType), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: util.MergeSchema(baseCriteriaSchema, map[string]*schema.Schema{
					"any_build": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any builds",
					},
					"selected_builds": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of build IDs",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied on which builds.",
		},
	})
}

var packBuildCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"any_build":       artifactoryCriteria["anyBuild"].(bool),
		"selected_builds": schema.NewSet(schema.HashString, artifactoryCriteria["selectedBuilds"].([]interface{})),
	}
}

var unpackBuildCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseWebhookCriteria) interface{} {
	return BuildWebhookCriteria{
		AnyBuild:            terraformCriteria["any_build"].(bool),
		SelectedBuilds:      util.CastToStringArr(terraformCriteria["selected_builds"].(*schema.Set).List()),
		BaseWebhookCriteria: baseCriteria,
	}
}

var buildCriteriaValidation = func(criteria map[string]interface{}) error {
	log.Print("[DEBUG] buildCriteriaValidation")

	anyBuild := criteria["any_build"].(bool)
	selectedBuilds := criteria["selected_builds"].(*schema.Set).List()

	if anyBuild == false && len(selectedBuilds) == 0 {
		return fmt.Errorf("selected_builds cannot be empty when any_build is false")
	}

	return nil
}
