package remote

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

type GenericRemoteRepo struct {
	RepositoryRemoteBaseParams
	PropagateQueryParams     bool `json:"propagateQueryParams"`
	RetrieveSha256FromServer bool `hcl:"retrieve_sha256_from_server" json:"retrieveSha256FromServer"`
}

var genericSchemaV3 = lo.Assign(
	map[string]*schema.Schema{
		"propagate_query_params": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set, if query params are included in the request to Artifactory, they will be passed on to the remote repository.",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, repository.GenericPackageType),
)

var genericSchemaV4 = lo.Assign(
	genericSchemaV3,
	map[string]*schema.Schema{
		"retrieve_sha256_from_server": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When set to `true`, Artifactory retrieves the SHA256 from the remote server if it is not cached in the remote repo.",
		},
	},
)

const currentGenericSchemaVersion = 4

var getSchemas = func(s map[string]*schema.Schema) map[int16]map[string]*schema.Schema {
	return map[int16]map[string]*schema.Schema{
		0: lo.Assign(
			baseSchemaV1,
			genericSchemaV3,
		),
		1: lo.Assign(
			baseSchemaV1,
			genericSchemaV3,
		),
		2: lo.Assign(
			baseSchemaV2,
			genericSchemaV3,
		),
		3: lo.Assign(
			baseSchemaV3,
			genericSchemaV3,
		),
		4: lo.Assign(
			baseSchemaV3,
			s,
		),
	}
}

var GenericSchemas = getSchemas(genericSchemaV4)

func ResourceArtifactoryRemoteGenericRepository() *schema.Resource {
	var unpackGenericRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: s}
		repo := GenericRemoteRepo{
			RepositoryRemoteBaseParams: UnpackBaseRemoteRepo(s, repository.GenericPackageType),
			PropagateQueryParams:       d.GetBool("propagate_query_params", false),
			RetrieveSha256FromServer:   d.GetBool("retrieve_sha256_from_server", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, repository.GenericPackageType)()
		if err != nil {
			return nil, err
		}

		return &GenericRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:        Rclass,
				PackageType:   repository.GenericPackageType,
				RepoLayoutRef: repoLayout.(string),
			},
		}, nil
	}

	resourceSchema := mkResourceSchema(
		GenericSchemas,
		packer.Default(GenericSchemas[currentGenericSchemaVersion]),
		unpackGenericRemoteRepo,
		constructor,
	)

	resourceSchema.Schema = GenericSchemas[currentGenericSchemaVersion]
	resourceSchema.SchemaVersion = currentGenericSchemaVersion
	resourceSchema.StateUpgraders = append(
		resourceSchema.StateUpgraders,
		schema.StateUpgrader{
			Type:    repository.Resource(GenericSchemas[3]).CoreConfigSchema().ImpliedType(),
			Upgrade: genericResourceStateUpgradeV3,
			Version: 3,
		},
	)

	return resourceSchema
}

func genericResourceStateUpgradeV3(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	rawState["retrieve_sha256_from_server"] = false
	if v, ok := rawState["property_sets"]; !ok || v == nil {
		rawState["property_sets"] = []string{}
	}

	return rawState, nil
}

var BasicSchema = func(packageType string) map[string]*schema.Schema {
	return lo.Assign(
		baseSchema,
		repository.RepoLayoutRefSchema(Rclass, packageType),
	)
}

func ResourceArtifactoryRemoteBasicRepository(packageType string) *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := repository.GetDefaultRepoLayoutRef(Rclass, packageType)()
		if err != nil {
			return nil, err
		}

		return &RepositoryRemoteBaseParams{
			PackageType:   packageType,
			Rclass:        Rclass,
			RepoLayoutRef: repoLayout.(string),
		}, nil
	}

	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackBaseRemoteRepo(data, packageType)
		return repo, repo.Id(), nil
	}

	basicSchema := BasicSchema(packageType)
	basicSchemas := GetSchemas(basicSchema)

	return mkResourceSchema(
		basicSchemas,
		packer.Default(basicSchemas[CurrentSchemaVersion]),
		unpack,
		constructor,
	)
}
