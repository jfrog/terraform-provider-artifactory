package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const rclass = "federated"

var MemberSchema = map[string]*schema.Schema{
	"member": {
		Type:     schema.TypeSet,
		Optional: true,
		Description: "The list of Federated members. If a Federated member receives a request that does not include the repository URL, it will " +
			"automatically be added with the combination of the configured base URL and `key` field value. " +
			"Note that each of the federated members will need to have a base URL set. Please follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)" +
			" to set up Federated repositories correctly.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"url": {
					Type:             schema.TypeString,
					Required:         true,
					Description:      "Full URL to ending with the repositoryName",
					ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
				},
				"enabled": {
					Type:     schema.TypeBool,
					Required: true,
					Description: "Represents the active state of the federated member. It is supported to " +
						"change the enabled status of my own member. The config will be updated on the other " +
						"federated members automatically.",
				},
			},
		},
	},
}
