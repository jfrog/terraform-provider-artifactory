package webhook

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type EmptyWebhookCriteria struct{}

var userWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return getBaseSchemaByVersion(webhookType, version, isCustom)
}

var packEmptyCriteria = func(artifactoryCriteria map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{}
}

var unpackEmptyCriteria = func(terraformCriteria map[string]interface{}, baseCriteria BaseCriteriaAPIModel) interface{} {
	return EmptyWebhookCriteria{}
}
