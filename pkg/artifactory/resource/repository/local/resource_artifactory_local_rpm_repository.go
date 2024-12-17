package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

var rpmSchema = utilsdk.MergeMaps(
	repository.PrimaryKeyPairRef,
	repository.SecondaryKeyPairRef,
	map[string]*schema.Schema{
		"yum_root_depth": {
			Type:             schema.TypeInt,
			Optional:         true,
			Default:          0,
			ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
			Description: "The depth, relative to the repository's root folder, where RPM metadata is created. " +
				"This is useful when your repository contains multiple RPM repositories under parallel hierarchies. " +
				"For example, if your RPMs are stored under 'fedora/linux/$releasever/$basearch', specify a depth of 4.",
		},
		"calculate_yum_metadata": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"enable_file_lists_indexing": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"yum_group_file_names": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "",
			ValidateDiagFunc: validator.CommaSeperatedList,
			Description: "A comma separated list of XML file names containing RPM group component definitions. Artifactory includes " +
				"the group definitions as part of the calculated RPM metadata, as well as automatically generating a " +
				"gzipped version of the group files, if required.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.RPMPackageType),
)

var RPMSchemas = GetSchemas(rpmSchema)

type RpmLocalRepositoryParams struct {
	RepositoryBaseParams
	repository.PrimaryKeyPairRefParam
	repository.SecondaryKeyPairRefParam
	RootDepth               int    `hcl:"yum_root_depth" json:"yumRootDepth"`
	CalculateYumMetadata    bool   `hcl:"calculate_yum_metadata" json:"calculateYumMetadata"`
	EnableFileListsIndexing bool   `hcl:"enable_file_lists_indexing" json:"enableFileListsIndexing"`
	GroupFileNames          string `hcl:"yum_group_file_names" json:"yumGroupFileNames"`
}

func UnpackLocalRpmRepository(data *schema.ResourceData, Rclass string) RpmLocalRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: data}
	return RpmLocalRepositoryParams{
		RepositoryBaseParams: UnpackBaseRepo(Rclass, data, repository.RPMPackageType),
		PrimaryKeyPairRefParam: repository.PrimaryKeyPairRefParam{
			PrimaryKeyPairRef: d.GetString("primary_keypair_ref", false),
		},
		SecondaryKeyPairRefParam: repository.SecondaryKeyPairRefParam{
			SecondaryKeyPairRef: d.GetString("secondary_keypair_ref", false),
		},
		RootDepth:               d.GetInt("yum_root_depth", false),
		CalculateYumMetadata:    d.GetBool("calculate_yum_metadata", false),
		EnableFileListsIndexing: d.GetBool("enable_file_lists_indexing", false),
		GroupFileNames:          d.GetString("yum_group_file_names", false),
	}
}

func ResourceArtifactoryLocalRpmRepository() *schema.Resource {
	unpackLocalRpmRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalRpmRepository(data, Rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &RpmLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: repository.RPMPackageType,
				Rclass:      Rclass,
			},
			RootDepth:               0,
			CalculateYumMetadata:    false,
			EnableFileListsIndexing: false,
			GroupFileNames:          "",
		}, nil
	}

	return repository.MkResourceSchema(
		RPMSchemas,
		packer.Default(RPMSchemas[CurrentSchemaVersion]),
		unpackLocalRpmRepository,
		constructor,
	)
}
