package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type GemsRemoteRepo struct {
	GenericRemoteRepo
	RepositoryCurationParams
}

var gemsSchemaV4 = lo.Assign(
	GenericSchemaV4,
	CurationRemoteRepoSchema,
)

const currentGemsSchemaVersion = 4

var GemsSchemas = GetGenericSchemas(gemsSchemaV4)

func ResourceArtifactoryRemoteGemsRepository() *schema.Resource {
	var unpackGemsRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := GemsRemoteRepo{
			GenericRemoteRepo: GenericRemoteRepo{
				RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.GemsPackageType),
				PropagateQueryParams:       d.GetBool("propagate_query_params", false),
				RetrieveSha256FromServer:   d.GetBool("retrieve_sha256_from_server", false),
			},
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.GemsPackageType)()
		if err != nil {
			return nil, err
		}

		return &GemsRemoteRepo{
			GenericRemoteRepo: GenericRemoteRepo{
				RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
					Rclass:        Rclass,
					PackageType:   repository.GemsPackageType,
					RepoLayoutRef: repoLayout.(string),
				},
			},
		}, nil
	}

	resourceSchema := mkResourceSchema(
		GemsSchemas,
		packer.Default(GemsSchemas[currentGemsSchemaVersion]),
		unpackGemsRemoteRepo,
		constructor,
	)

	resourceSchema.Schema = GemsSchemas[currentGemsSchemaVersion]
	resourceSchema.SchemaVersion = currentGemsSchemaVersion
	resourceSchema.StateUpgraders = append(
		resourceSchema.StateUpgraders,
		schema.StateUpgrader{
			Type:    repository.Resource(GenericSchemas[3]).CoreConfigSchema().ImpliedType(),
			Upgrade: GenericResourceStateUpgradeV3,
			Version: 3,
		},
	)

	return resourceSchema
}
