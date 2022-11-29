package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func ResourceArtifactoryRemoteJavaRepository(repoType string, suppressPom bool) *schema.Resource {
	javaRemoteSchema := getJavaRemoteSchema(repoType, suppressPom)

	var unpackJavaRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackJavaRemoteRepo(data, repoType)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &JavaRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: repoType,
			},
			SuppressPomConsistencyChecks: suppressPom,
		}, nil
	}

	return repository.MkResourceSchema(javaRemoteSchema, packer.Default(javaRemoteSchema), unpackJavaRemoteRepo, constructor)
}
