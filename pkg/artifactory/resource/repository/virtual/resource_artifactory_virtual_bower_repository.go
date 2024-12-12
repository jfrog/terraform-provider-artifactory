package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

var bowerSchema = lo.Assign(
	externalDependenciesSchema,
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.BowerPackageType),
)

var BowerSchemas = GetSchemas(bowerSchema)

func ResourceArtifactoryVirtualBowerRepository() *schema.Resource {
	var unpackBowerVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		repo := unpackExternalDependenciesVirtualRepository(s, repository.BowerPackageType)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &ExternalDependenciesVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.BowerPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		BowerSchemas,
		packer.Default(BowerSchemas[CurrentSchemaVersion]),
		unpackBowerVirtualRepository,
		constructor,
	)
}
