package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var helmOCISchema = lo.Assign(
	map[string]*schema.Schema{
		"resolve_oci_tags_by_timestamp": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When enabled, in cases where the same OCI tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.HelmOCIPackageType),
)

var HelmOCISchemas = GetSchemas(helmOCISchema)

type HelmOciVirtualRepositoryParams struct {
	RepositoryBaseParams
	ResolveOCITagsByTimestamp bool `hcl:"resolve_oci_tags_by_timestamp" json:"resolveDockerTagsByTimestamp"`
}

func ResourceArtifactoryVirtualHelmOciRepository() *schema.Resource {
	unpackVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := HelmOciVirtualRepositoryParams{
			RepositoryBaseParams:      UnpackBaseVirtRepo(data, repository.HelmOCIPackageType),
			ResolveOCITagsByTimestamp: d.GetBool("resolve_oci_tags_by_timestamp", false),
		}

		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &HelmOciVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.HelmOCIPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		HelmOCISchemas,
		packer.Default(HelmOCISchemas[CurrentSchemaVersion]),
		unpackVirtualRepository,
		constructor,
	)
}
