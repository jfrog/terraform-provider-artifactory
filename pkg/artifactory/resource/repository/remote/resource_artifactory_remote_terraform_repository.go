package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type TerraformRemoteRepo struct {
	RepositoryRemoteBaseParams
	TerraformRegistryUrl  string `json:"terraformRegistryUrl"`
	TerraformProvidersUrl string `json:"terraformProvidersUrl"`
}

const TerraformPackageType = "terraform"

var TerraformRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return util.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"terraform_registry_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Default:      "https://registry.terraform.io",
				Description: "The base URL of the registry API. When using Smart Remote Repositories, set the URL to" +
					" <base_Artifactory_URL>/api/terraform/repokey. Default value in UI is https://registry.terraform.io",
			},
			"terraform_providers_url": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
				Default:      "https://releases.hashicorp.com",
				Description: "The base URL of the Provider's storage API. When using Smart remote repositories, set " +
					"the URL to <base_Artifactory_URL>/api/terraform/repokey/providers. Default value in UI is https://releases.hashicorp.com",
			},
		},
		repository.RepoLayoutRefSchema(rclass, TerraformPackageType),
	)
}

func ResourceArtifactoryRemoteTerraformRepository() *schema.Resource {
	var unpackTerraformRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}
		repo := TerraformRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, TerraformPackageType),
			TerraformRegistryUrl:       d.GetString("terraform_registry_url", false),
			TerraformProvidersUrl:      d.GetString("terraform_providers_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &TerraformRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      rclass,
				PackageType: TerraformPackageType,
			},
		}, nil
	}

	terraformSchema := TerraformRemoteSchema(true)

	return mkResourceSchema(terraformSchema, packer.Default(terraformSchema), unpackTerraformRemoteRepo, constructor)
}
