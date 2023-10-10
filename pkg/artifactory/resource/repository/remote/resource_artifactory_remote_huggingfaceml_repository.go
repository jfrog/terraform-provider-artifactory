package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const HuggingFacePackageType = "huggingfaceml"

type HuggingFaceRepo struct {
	RepositoryRemoteBaseParams
}

var HuggingFaceSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		BaseRemoteRepoSchema(isResource),
		map[string]*schema.Schema{
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "https://huggingface.co",
				Description: "The remote repo URL. Default to 'https://huggingface.co'",
			},
		},
		repository.RepoLayoutRefSchema(rclass, HuggingFacePackageType),
	)
}

func ResourceArtifactoryRemoteHuggingFaceRepository() *schema.Resource {
	var unpackRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		repo := HuggingFaceRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, HuggingFacePackageType),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(rclass, HuggingFacePackageType)()
		if err != nil {
			return nil, err
		}

		return &HuggingFaceRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        rclass,
				PackageType:   HuggingFacePackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	huggingFaceSchema := HuggingFaceSchema(true)

	return mkResourceSchema(huggingFaceSchema, packer.Default(huggingFaceSchema), unpackRepo, constructor)
}
