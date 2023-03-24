package local

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

const rpmPackageType = "rpm"

var RpmLocalSchema = util.MergeMaps(
	BaseLocalRepoSchema,
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
		"primary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Primary keypair used to sign artifacts.",
		},
		"secondary_keypair_ref": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Secondary keypair used to sign artifacts.",
		},
	},
	repository.RepoLayoutRefSchema(rclass, rpmPackageType),
)

type RpmLocalRepositoryParams struct {
	RepositoryBaseParams
	RootDepth               int    `hcl:"yum_root_depth" json:"yumRootDepth"`
	CalculateYumMetadata    bool   `hcl:"calculate_yum_metadata" json:"calculateYumMetadata"`
	EnableFileListsIndexing bool   `hcl:"enable_file_lists_indexing" json:"enableFileListsIndexing"`
	GroupFileNames          string `hcl:"yum_group_file_names" json:"yumGroupFileNames"`
	PrimaryKeyPairRef       string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
	SecondaryKeyPairRef     string `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef"`
}

func UnpackLocalRpmRepository(data *schema.ResourceData, rclass string) RpmLocalRepositoryParams {
	d := &util.ResourceData{ResourceData: data}
	return RpmLocalRepositoryParams{
		RepositoryBaseParams:    UnpackBaseRepo(rclass, data, rpmPackageType),
		RootDepth:               d.GetInt("yum_root_depth", false),
		CalculateYumMetadata:    d.GetBool("calculate_yum_metadata", false),
		EnableFileListsIndexing: d.GetBool("enable_file_lists_indexing", false),
		GroupFileNames:          d.GetString("yum_group_file_names", false),
		PrimaryKeyPairRef:       d.GetString("primary_keypair_ref", false),
		SecondaryKeyPairRef:     d.GetString("secondary_keypair_ref", false),
	}
}

func ResourceArtifactoryLocalRpmRepository() *schema.Resource {

	unpackLocalRpmRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackLocalRpmRepository(data, rclass)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &RpmLocalRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				PackageType: rpmPackageType,
				Rclass:      rclass,
			},
			RootDepth:               0,
			CalculateYumMetadata:    false,
			EnableFileListsIndexing: false,
			GroupFileNames:          "",
		}, nil
	}

	return repository.MkResourceSchema(RpmLocalSchema, packer.Default(RpmLocalSchema), unpackLocalRpmRepository, constructor)
}
