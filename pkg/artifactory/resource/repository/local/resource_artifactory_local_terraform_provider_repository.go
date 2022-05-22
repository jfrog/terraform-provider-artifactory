package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

var terraformProviderLocalSchema = util.MergeSchema(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
		"registry_type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "provider",
			Description: "The Terraform registry in Artifactory allows you to create dedicated repositories for the " +
				"unique Terraform component Provider.",
			ValidateDiagFunc: validator.StringInSlice(true, "provider"),
		},
	},
	repository.RepoLayoutRefSchema("local", "terraform_provider"),
)

func ResourceArtifactoryLocalTerraformProviderRepository() *schema.Resource {

	type TerraformLocalRepo struct {
		LocalRepositoryBaseParams
		TerraformType string `json:"terraformType"`
	}

	var unPackLocalTerraformProviderRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := TerraformLocalRepo{
			LocalRepositoryBaseParams: UnpackBaseRepo("local", data, "terraform"),
			TerraformType:             d.GetString("registry_type", false),
		}

		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(terraformProviderLocalSchema, repository.DefaultPacker(terraformProviderLocalSchema), unPackLocalTerraformProviderRepository, func() interface{} {
		return &TerraformLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "terraform",
				Rclass:      "local",
			},
		}
	})
}
