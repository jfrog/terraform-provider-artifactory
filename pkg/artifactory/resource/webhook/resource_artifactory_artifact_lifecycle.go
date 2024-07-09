package webhook

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var artifactLifecycleWebhookSchema = func(webhookType string, version int, isCustom bool) map[string]*schema.Schema {
	return getBaseSchemaByVersion(webhookType, version, isCustom)
}
