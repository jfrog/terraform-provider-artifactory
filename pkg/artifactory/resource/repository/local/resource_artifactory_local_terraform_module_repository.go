package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

var terraformModuleLocalSchema = util.MergeSchema(
	BaseLocalRepoSchema,
	map[string]*schema.Schema{
		"registry_type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "module",
			Description: "The Terraform registry in Artifactory allows you to create dedicated repositories for the " +
				"unique Terraform component Module.",
			ValidateDiagFunc: validator.StringInSlice(true, "module"),
		},
	},
	repository.RepoLayoutRefSchema("local", "terraform_module"),
)

func ResourceArtifactoryLocalTerraformModuleRepository() *schema.Resource {

	type TerraformLocalRepo struct {
		LocalRepositoryBaseParams
		TerraformType string `json:"terraformType"`
	}

	var unPackLocalTerraformModuleRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := TerraformLocalRepo{
			LocalRepositoryBaseParams: UnpackBaseRepo("local", data, "terraform"),
			TerraformType:             d.GetString("registry_type", false),
		}

		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(terraformModuleLocalSchema, repository.DefaultPacker(terraformModuleLocalSchema), unPackLocalTerraformModuleRepository, func() interface{} {
		return &TerraformLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				PackageType: "terraform",
				Rclass:      "local",
			},
		}
	})
}
