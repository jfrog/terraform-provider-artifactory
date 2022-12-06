package federated

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
)

const rclass = "federated"

var RepoTypesLikeGeneric = []string{
	"bower",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"helm",
	"npm",
	"nuget",
	"opkg",
	"puppet",
	"pypi",
	"rpm",
	"swift",
	"terraform_module",
	"terraform_provider",
	"vagrant",
}

type Member struct {
	Url     string `hcl:"url" json:"url"`
	Enabled bool   `hcl:"enabled" json:"enabled"`
}

var memberSchema = map[string]*schema.Schema{
	"member": {
		Type:     schema.TypeSet,
		Required: true,
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

var unpackMembers = func(data *schema.ResourceData) []Member {
	d := &util.ResourceData{ResourceData: data}
	var members []Member

	if v, ok := d.GetOkExists("member"); ok {
		federatedMembers := v.(*schema.Set).List()
		if len(federatedMembers) == 0 {
			return members
		}

		for _, federatedMember := range federatedMembers {
			id := federatedMember.(map[string]interface{})

			member := Member{
				Url:     id["url"].(string),
				Enabled: id["enabled"].(bool),
			}
			members = append(members, member)
		}
	}
	return members
}

func packMembers(members []Member, d *schema.ResourceData) error {
	setValue := util.MkLens(d)

	var federatedMembers []interface{}

	for _, member := range members {
		federatedMember := map[string]interface{}{
			"url":     member.Url,
			"enabled": member.Enabled,
		}

		federatedMembers = append(federatedMembers, federatedMember)
	}

	errors := setValue("member", federatedMembers)
	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed saving members to state %q", errors)
	}

	return nil
}
