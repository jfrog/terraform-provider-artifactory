package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/packer"
)

func ResourceArtifactoryRemoteJavaRepository(packageType string, suppressPom bool) *schema.Resource {
	javaRemoteSchema := JavaRemoteSchema(true, packageType, suppressPom)

	var unpackJavaRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackJavaRemoteRepo(data, packageType)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &JavaRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      rclass,
				PackageType: packageType,
			},
			SuppressPomConsistencyChecks: suppressPom,
		}, nil
	}

	return mkResourceSchema(javaRemoteSchema, packer.Default(javaRemoteSchema), unpackJavaRemoteRepo, constructor)
}
