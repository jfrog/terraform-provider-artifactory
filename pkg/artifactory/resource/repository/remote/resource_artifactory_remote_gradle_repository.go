package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type GradleRemoteRepo struct {
	RepositoryCurationParams
	JavaRemoteRepo
}

func ResourceArtifactoryRemoteGradleRepository() *schema.Resource {
	gradleSchema := lo.Assign(
		CurationRemoteRepoSchema,
		JavaSchema(repository.GradlePackageType, true),
	)

	gradleSchemas := GetSchemas(gradleSchema)

	var unpackGradleRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := GradleRemoteRepo{
			JavaRemoteRepo: UnpackJavaRemoteRepo(data, repository.GradlePackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &GradleRemoteRepo{
			JavaRemoteRepo: JavaRemoteRepo{
				RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
					Rclass:      Rclass,
					PackageType: repository.GradlePackageType,
				},
				SuppressPomConsistencyChecks: true,
			},
		}, nil
	}

	return mkResourceSchema(
		gradleSchemas,
		packer.Default(gradleSchemas[CurrentSchemaVersion]),
		unpackGradleRemoteRepo,
		constructor,
	)
}
