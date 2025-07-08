package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type HexVirtualRepositoryParams struct {
	RepositoryBaseParams
	PrimaryKeyPairRef string `hcl:"hex_primary_keypair_ref" json:"primaryKeyPairRef"`
}

func ResourceArtifactoryVirtualHexRepository() *schema.Resource {
	var hexVirtualSchema = lo.Assign(
		BaseSchemaV1,
		map[string]*schema.Schema{
			"hex_primary_keypair_ref": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Primary keypair used to sign artifacts.",
			},
		},
	)

	constructor := func() (interface{}, error) {
		return &HexVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.HexPackageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := HexVirtualRepositoryParams{
			RepositoryBaseParams: UnpackBaseVirtRepo(data, repository.HexPackageType),
			PrimaryKeyPairRef:    d.GetString("hex_primary_keypair_ref", false),
		}
		return repo, repo.Id(), nil
	}

	return repository.MkResourceSchema(
		map[int16]map[string]*schema.Schema{
			0: hexVirtualSchema,
			1: hexVirtualSchema,
		},
		packer.Default(hexVirtualSchema),
		unpack,
		constructor,
	)
}
