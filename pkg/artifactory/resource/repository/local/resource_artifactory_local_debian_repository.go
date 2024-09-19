package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var debianSchema = lo.Assign(
	repository.PrimaryKeyPairRef,
	repository.SecondaryKeyPairRef,
	map[string]*schema.Schema{
		"trivial_layout": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, the repository will use the deprecated trivial layout.",
			Deprecated:  "You shouldn't be using this",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, repository.DebianPackageType),
	repository.CompressionFormats,
)

var DebianSchemas = GetSchemas(debianSchema)

type DebianLocalRepositoryParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
	repository.SecondaryKeyPairRefParam
	TrivialLayout           bool     `hcl:"trivial_layout" json:"debianTrivialLayout"`
	IndexCompressionFormats []string `hcl:"index_compression_formats" json:"optionalIndexCompressionFormats,omitempty"`
}

func UnpackLocalDebianRepository(data *schema.ResourceData, Rclass string) DebianLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return DebianLocalRepositoryParams{
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
		},
		SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
			SecondaryKeyPairRef: d.GetString("secondary_keypair_ref", false),
		},
		RepositoryBaseParams:    UnpackBaseRepo(Rclass, data, repository.DebianPackageType),
		TrivialLayout:           d.GetBool("trivial_layout", false),
		IndexCompressionFormats: d.GetSet("index_compression_formats"),
	}
}

func ResourceArtifactoryLocalDebianRepository() *schema.Resource {

	var unpackLocalDebianRepository = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalDebianRepository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DebianLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.DebianPackageType,
				Rclass:      Rclass,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		DebianSchemas,
		packer.Default(DebianSchemas[CurrentSchemaVersion]),
		unpackLocalDebianRepository,
		constructor,
	)
}
