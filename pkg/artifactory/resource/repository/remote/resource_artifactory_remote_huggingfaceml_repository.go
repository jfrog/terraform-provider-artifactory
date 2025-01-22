package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type HuggingFaceRepo struct {
	RepositoryRemoteBaseParams
	RepositoryCurationParams
}

var HuggingFaceSchema = lo.Assign(
	BaseSchema,
	CurationRemoteRepoSchema,
	map[string]*schema.Schema{
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "https://huggingface.co",
			Description: "The remote repo URL. Default to 'https://huggingface.co'",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.HuggingFacePackageType),
)

var HuggingFaceSchemas = GetSchemas(HuggingFaceSchema)

func ResourceArtifactoryRemoteHuggingFaceRepository() *schema.Resource {
	var unpackRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := HuggingFaceRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.HuggingFacePackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.HuggingFacePackageType)
		if err != nil {
			return nil, err
		}

		return &HuggingFaceRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.HuggingFacePackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	return mkResourceSchema(
		HuggingFaceSchemas,
		packer.Default(HuggingFaceSchemas[CurrentSchemaVersion]),
		unpackRepo,
		constructor,
	)
}
