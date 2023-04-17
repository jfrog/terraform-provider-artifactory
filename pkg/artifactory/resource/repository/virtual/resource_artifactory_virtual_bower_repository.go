package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const BowerPackageType = "bower"

var BowerVirtualSchema = util.MergeMaps(
	BaseVirtualRepoSchema,
	externalDependenciesSchema,
	repository.RepoLayoutRefSchema(Rclass, BowerPackageType),
)

func ResourceArtifactoryVirtualBowerRepository() *schema.Resource {
	var unpackBowerVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		repo := unpackExternalDependenciesVirtualRepository(s, BowerPackageType)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &ExternalDependenciesVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: BowerPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		BowerVirtualSchema,
		packer.Default(BowerVirtualSchema),
		unpackBowerVirtualRepository,
		constructor,
	)
}
