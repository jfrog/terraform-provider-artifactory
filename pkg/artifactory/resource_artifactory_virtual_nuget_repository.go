package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func resourceArtifactoryNugetVirtualRepository() *schema.Resource {

	const packageType = "nuget"

	var nugetVirtualSchema = utils.MergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{
		"force_nuget_authentication": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Force basic authentication credentials in order to use this repository.",
		},
	}, repoLayoutRefSchema("virtual", packageType))

	type NugetVirtualRepositoryParams struct {
		VirtualRepositoryBaseParams
		ForceNugetAuthentication bool `json:"forceNugetAuthentication"`
	}

	var unpackNugetVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utils.ResourceData{s}

		repo := NugetVirtualRepositoryParams{
			VirtualRepositoryBaseParams: unpackBaseVirtRepo(s, packageType),
			ForceNugetAuthentication:    d.GetBool("force_nuget_authentication", false),
		}
		repo.PackageType = packageType
		return &repo, repo.Key, nil
	}

	return mkResourceSchema(nugetVirtualSchema, defaultPacker(nugetVirtualSchema), unpackNugetVirtualRepository, func() interface{} {
		return &NugetVirtualRepositoryParams{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: packageType,
			},
		}
	})
}
