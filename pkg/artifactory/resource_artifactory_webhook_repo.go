package artifactory

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

type RepoWebhookCriteria struct {
	BaseWebhookCriteria
	AnyLocal  bool     `json:"anyLocal"`
	AnyRemote bool     `json:"anyRemote"`
	RepoKeys  []string `json:"repoKeys"`
}

var repoWebhookSchema = func(webhookType string) map[string]*schema.Schema {
	return utils.MergeSchema(baseWebhookBaseSchema(webhookType), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utils.MergeSchema(baseCriteriaSchema, map[string]*schema.Schema{
					"any_local": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any local repositories",
					},
					"any_remote": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any remote repositories",
					},
					"repo_keys": {
						Type:        schema.TypeSet,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Trigger on this list of repository keys",
					},
				}),
			},
			Description: "Specifies where the webhook will be applied on which repositories.",
		},
	})
}

var packRepoCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"any_local":  artifactoryCriteria["anyLocal"].(bool),
		"any_remote": artifactoryCriteria["anyRemote"].(bool),
		"repo_keys":  schema.NewSet(schema.HashString, artifactoryCriteria["repoKeys"].([]interface{})),
	}
}

var unpackRepoCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseWebhookCriteria) interface{} {
	return RepoWebhookCriteria{
		AnyLocal:            terraformCriteria["any_local"].(bool),
		AnyRemote:           terraformCriteria["any_remote"].(bool),
		RepoKeys:            utils.CastToStringArr(terraformCriteria["repo_keys"].(*schema.Set).List()),
		BaseWebhookCriteria: baseCriteria,
	}
}

var repoCriteriaValidation = func(criteria map[string]interface{}) error {
	log.Print("[DEBUG] repoCriteriaValidation")

	anyLocal := criteria["any_local"].(bool)
	anyRemote := criteria["any_remote"].(bool)
	repoKeys := criteria["repo_keys"].(*schema.Set).List()

	if (anyLocal == false && anyRemote == false) && len(repoKeys) == 0 {
		return fmt.Errorf("repo_keys cannot be empty when both any_local and any_remote are false")
	}

	return nil
}
