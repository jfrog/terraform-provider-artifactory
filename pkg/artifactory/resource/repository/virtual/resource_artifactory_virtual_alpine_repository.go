package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const AlpinePackageType = "alpine"

var AlpineVirtualSchema = util.MergeMaps(
	BaseVirtualRepoSchema,
	RetrievalCachePeriodSecondsSchema,
	map[string]*schema.Schema{
		"primary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Primary keypair used to sign artifacts. Default value is empty.",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, AlpinePackageType))

func ResourceArtifactoryVirtualAlpineRepository() *schema.Resource {
	type AlpineVirtualRepositoryParams struct {
		RepositoryBaseParamsWithRetrievalCachePeriodSecs
		PrimaryKeyPairRef string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
	}

	var unpackAlpineVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}

		repo := AlpineVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s, AlpinePackageType),
			PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
		}
		repo.PackageType = AlpinePackageType
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &AlpineVirtualRepositoryParams{
			RepositoryBaseParamsWithRetrievalCachePeriodSecs: RepositoryBaseParamsWithRetrievalCachePeriodSecs{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      Rclass,
					PackageType: AlpinePackageType,
				},
			},
		}, nil
	}

	return repository.MkResourceSchema(
		AlpineVirtualSchema,
		packer.Default(AlpineVirtualSchema),
		unpackAlpineVirtualRepository,
		constructor,
	)
}
