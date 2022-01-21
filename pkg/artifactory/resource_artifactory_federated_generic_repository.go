package artifactory

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceArtifactoryFederatedGenericRepository(repoType string) *schema.Resource {
	var federatedSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
		"member": {
			Type:     schema.TypeSet,
			Optional: true,
			// Computed: true,
			Description: "The list of Federated members. If a Federated member receives a request that does not include the repository URL, it will " +
				"automatically be added with the combination of the configured base URL and `key` field value. " +
				"Note that each of the federated members will need to have a base URL set. PLease follow the [instruction](https://www.jfrog.com/confluence/display/JFROG/Working+with+Federated+Repositories#WorkingwithFederatedRepositories-SettingUpaFederatedRepository)" +
				" to set up Federated repositories correctly.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"url": {
						Type:        schema.TypeString,
						Optional:    true,
						Computed:    true,
						Description: "Full URL to ending with the repositoryName",
						ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
					},
					"enabled": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Represents the active state of the federated member. It is supported to " +
							"change the enabled status of my own member. The config will be updated on the other " +
							"federated members automatically.",
					},
				},
			},
		},
	})

	type Member struct {
		Url     string `hcl:"url" json:"url"`
		Enabled bool   `hcl:"enabled" json:"enabled"`
	}

	type FederatedRepositoryParams struct {
		LocalRepositoryBaseParams
		Members []Member `hcl:"member" json:"members"`
	}

	var unpackMembers = func(data *schema.ResourceData) []Member {
		d := &ResourceData{data}

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

	var unPackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := FederatedRepositoryParams{
			LocalRepositoryBaseParams: unpackBaseRepo("federated", data, repoType),
			Members:                   unpackMembers(data),
		}

		return repo, repo.Id(), nil
	}

	var packMembers = func(repo interface{}, d *schema.ResourceData) error {
		setValue := mkLens(d)

		var federatedMembers []interface{}

		members := repo.(*FederatedRepositoryParams).Members
		for _, member := range members {
			federatedMember := map[string]interface{}{
				"url":         	member.Url,
				"enabled":  	member.Enabled,
			}

			federatedMembers = append(federatedMembers, federatedMember)
		}

		errors := setValue("member", federatedMembers)
		log.Printf("ResourceData.member: %v", d.Get("member"))

		if errors != nil && len(errors) > 0 {
			return fmt.Errorf("failed saving members to state %q", errors)
		}

		return nil
	}

	packer := composePacker(
		universalPack(ignoreHclPredicate("class", "rclass", "member")),
		packMembers,
	)

	constructor := func() interface{} {
		return &FederatedRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: repoType,
				Rclass:      "federated", //TODO: the value is not in the rest call, but it is in the schema
			},
		}
	}

	return mkResourceSchema(federatedSchema, packer, unPackFederatedRepository, constructor)
}
