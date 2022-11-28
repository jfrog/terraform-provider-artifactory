package federated

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository/local"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryFederatedGenericRepository(repoType string) *schema.Resource {
	localRepoSchema := local.GetSchemaByRepoType(repoType)

	var federatedSchema = util.MergeMaps(localRepoSchema, map[string]*schema.Schema{
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
	}, repository.RepoLayoutRefSchema("federated", repoType))

	type Member struct {
		Url     string `hcl:"url" json:"url"`
		Enabled bool   `hcl:"enabled" json:"enabled"`
	}

	type FederatedRepositoryParams struct {
		local.RepositoryBaseParams
		Members []Member `hcl:"member" json:"members"`
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

	var unpackFederatedRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := FederatedRepositoryParams{
			RepositoryBaseParams: local.UnpackBaseRepo("federated", data, repoType),
			Members:              unpackMembers(data),
		}
		// terraformType could be `module` or `provider`, repoType names we use are `terraform_module` and `terraform_provider`
		// We need to remove the `terraform_` from the string.
		repo.TerraformType = strings.ReplaceAll(repoType, "terraform_", "")

		return repo, repo.Id(), nil
	}

	var packMembers = func(repo interface{}, d *schema.ResourceData) error {
		setValue := util.MkLens(d)

		var federatedMembers []interface{}

		members := repo.(*FederatedRepositoryParams).Members
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

	pkr := packer.Compose(
		packer.Universal(
			predicate.Ignore("class", "rclass", "member", "terraform_type"),
		),
		packMembers,
	)

	constructor := func() (interface{}, error) {
		return &FederatedRepositoryParams{
			RepositoryBaseParams: local.RepositoryBaseParams{
				PackageType: local.GetPackageType(repoType),
				Rclass:      "federated",
			},
		}, nil
	}

	return repository.MkResourceSchema(federatedSchema, pkr, unpackFederatedRepository, constructor)
}
