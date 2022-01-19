package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceArtifactoryFederatedGenericRepository(repoType string) *schema.Resource {
	var federatedSchema = mergeSchema(baseLocalRepoSchema, map[string]*schema.Schema{
		"member": {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
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
						//ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
					},
					"enabled": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Represents the active state of the federated member. It is supported to " +
							"change the enabled status of my own member. The config will be updated on the other " +
							"federated members automatically.",
						//ValidateDiagFunc: validation.ToDiagFunc(validation.NoZeroValues),
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
		Member []Member `hcl:"member" json:"members"`
	}

	var unpackMembers = func(data *schema.ResourceData) []Member {
		d := &ResourceData{data}

		var members []Member

		if v, ok := d.GetOkExists("member"); ok {
			federatedMembers := v.(*schema.Set).List() // exception here if I use TypeList in the schema
			fmt.Println(federatedMembers)              // debug
			if len(federatedMembers) == 0 {
				return members // return empty array
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
			Member:                    unpackMembers(data),
		}

		return repo, repo.Id(), nil
	}

	//TODO: this function has a problem, universalPack works only with []string, not []interfece{}, a lot of dependencies

	var packMembers = func(repo interface{}, d *schema.ResourceData) error {
		setValue := mkLens(d)

		var federatedMembers []interface{}

		//TODO:  try to read data from repo
		values := lookup(repo)

		for _, value := range values {
			fmt.Println(value)

			//federatedMember := map[string]interface{}{
			//	"url":         	member.Url,
			//	"enabled":  	member.Enabled,
			//}

			//federatedMembers = append(federatedMembers, federatedMember)
		}

		log.Printf("[TRACE] %+v\n", federatedMembers)
		errors := setValue("member", federatedMembers)

		if errors != nil && len(errors) > 0 {
			return fmt.Errorf("failed saving members to state %q", errors)
		}

		return nil
	}

	//type PackFunc func(repo interface{}, d *schema.ResourceData) error

	//var members []Member
	//var repo = FederatedRepositoryParams{}

	packer := composePacker(
		universalPack(ignoreHclPredicate("member")),
		packMembers,
	)
	return mkResourceSchema(federatedSchema, packer, unPackFederatedRepository, func() interface{} {
		return &FederatedRepositoryParams{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: repoType,
				Rclass:      "federated", //TODO: the value is not in the rest call, but it is in the schema
			},
		}
	})
}
