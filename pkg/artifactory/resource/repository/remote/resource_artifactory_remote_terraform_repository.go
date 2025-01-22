package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type TerraformRemoteRepo struct {
	RepositoryRemoteBaseParams
	TerraformRegistryUrl  string `json:"terraformRegistryUrl"`
	TerraformProvidersUrl string `json:"terraformProvidersUrl"`
}

var TerraformSchema = lo.Assign(
	BaseSchema,
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
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.TerraformPackageType),
)

var TerraformSchemas = GetSchemas(TerraformSchema)

func ResourceArtifactoryRemoteTerraformRepository() *schema.Resource {
	var unpackTerraformRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := TerraformRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.TerraformPackageType),
			TerraformRegistryUrl:       d.GetString("terraform_registry_url", false),
			TerraformProvidersUrl:      d.GetString("terraform_providers_url", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &TerraformRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      Rclass,
				PackageType: repository.TerraformPackageType,
			},
		}, nil
	}

	return mkResourceSchema(
		TerraformSchemas,
		packer.Default(TerraformSchemas[CurrentSchemaVersion]),
		unpackTerraformRemoteRepo,
		constructor,
	)
}
