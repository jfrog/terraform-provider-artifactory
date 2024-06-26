package webhook

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

type RepoWebhookCriteria struct {
	BaseWebhookCriteria
	AnyLocal     bool     `json:"anyLocal"`
	AnyRemote    bool     `json:"anyRemote"`
	AnyFederated bool     `json:"anyFederated"`
	RepoKeys     []string `json:"repoKeys"`
}

var repoWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(getBaseSchemaByVersion(webhookType, version, isCustom), map[string]*schema.Schema{
		"criteria": {
			Type:     schema.TypeSet,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: utilsdk.MergeMaps(baseCriteriaSchema, map[string]*schema.Schema{
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
					"any_federated": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Trigger on any federated repositories",
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
	criteria := map[string]interface{}{
		"any_local":     artifactoryCriteria["anyLocal"].(bool),
		"any_remote":    artifactoryCriteria["anyRemote"].(bool),
		"any_federated": false,
		"repo_keys":     schema.NewSet(schema.HashString, artifactoryCriteria["repoKeys"].([]interface{})),
	}

	if v, ok := artifactoryCriteria["anyFederated"]; ok {
		criteria["any_federated"] = v.(bool)
	}

	return criteria
}

var unpackRepoCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseWebhookCriteria) interface{} {
	return RepoWebhookCriteria{
		AnyLocal:            terraformCriteria["any_local"].(bool),
		AnyRemote:           terraformCriteria["any_remote"].(bool),
		AnyFederated:        terraformCriteria["any_federated"].(bool),
		RepoKeys:            utilsdk.CastToStringArr(terraformCriteria["repo_keys"].(*schema.Set).List()),
		BaseWebhookCriteria: baseCriteria,
	}
}

var repoCriteriaValidation = func(ctx context.Context, criteria map[string]interface{}) error {
	tflog.Debug(ctx, "repoCriteriaValidation")

	anyLocal := criteria["any_local"].(bool)
	anyRemote := criteria["any_remote"].(bool)
	anyFederated := criteria["any_federated"].(bool)
	repoKeys := criteria["repo_keys"].(*schema.Set).List()

	if (!anyLocal && !anyRemote && !anyFederated) && len(repoKeys) == 0 {
		return fmt.Errorf("repo_keys cannot be empty when any_local, any_remote, and any_federated are false")
	}

	return nil
}
