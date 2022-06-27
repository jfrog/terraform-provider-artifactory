package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

type TerraformRemoteRepo struct {
	RemoteRepositoryBaseParams
	TerraformRegistryUrl  string `hcl:"terraform_registry_url" json:"terraformRegistryUrl"`
	TerraformProvidersUrl string `hcl:"terraform_providers_url" json:"terraformProvidersUrl"`
}

func ResourceArtifactoryRemoteTerraformRepository() *schema.Resource {
	const packageType = "terraform"

	var terraformRemoteSchema = util.MergeMaps(BaseRemoteRepoSchema, map[string]*schema.Schema{
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
	}, repository.RepoLayoutRefSchema("remote", packageType))

	var unpackTerraformRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{s}
		repo := TerraformRemoteRepo{
			RemoteRepositoryBaseParams: UnpackBaseRemoteRepo(s, packageType),
			TerraformRegistryUrl:       d.GetString("terraform_registry_url", false),
			TerraformProvidersUrl:      d.GetString("terraform_providers_url", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(terraformRemoteSchema, packer.Default(terraformRemoteSchema), unpackTerraformRemoteRepo, func() interface{} {
		return &TerraformRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}
	})
}
