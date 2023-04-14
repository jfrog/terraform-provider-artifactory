package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const NugetPackageType = "nuget"

var NugetVirtualSchema = util.MergeMaps(BaseVirtualRepoSchema, map[string]*schema.Schema{
	"force_nuget_authentication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "If set, user authentication is required when accessing the repository. An anonymous request will display an HTTP 401 error. This is also enforced when aggregated repositories support anonymous requests.",
	},
}, repository.RepoLayoutRefSchema(Rclass, NugetPackageType))

func ResourceArtifactoryVirtualNugetRepository() *schema.Resource {

	type NugetVirtualRepositoryParams struct {
		RepositoryBaseParams
		ForceNugetAuthentication bool `json:"forceNugetAuthentication"`
	}

	var unpackNugetVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}

		repo := NugetVirtualRepositoryParams{
			RepositoryBaseParams:     UnpackBaseVirtRepo(s, NugetPackageType),
			ForceNugetAuthentication: d.GetBool("force_nuget_authentication", false),
		}
		repo.PackageType = NugetPackageType
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &NugetVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: NugetPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		NugetVirtualSchema,
		packer.Default(NugetVirtualSchema),
		unpackNugetVirtualRepository,
		constructor,
	)
}
