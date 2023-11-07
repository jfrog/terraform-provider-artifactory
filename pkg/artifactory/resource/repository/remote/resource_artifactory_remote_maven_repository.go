package remote

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const MavenPackageType = "maven"

type MavenRemoteRepo struct {
	JavaRemoteRepo
	RepositoryCurationParams
}

var MavenRemoteSchema = func(isResource bool) map[string]*schema.Schema {
	return utilsdk.MergeMaps(
		JavaRemoteSchema(isResource, MavenPackageType, false),
		CurationRemoteRepoSchema,
	)
}

func ResourceArtifactoryRemoteMavenRepository() *schema.Resource {
	mavenRemoteSchema := MavenRemoteSchema(true)

	var unpackMavenRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := MavenRemoteRepo{
			JavaRemoteRepo: UnpackJavaRemoteRepo(data, MavenPackageType),
			RepositoryCurationParams: RepositoryCurationParams{
				Curated: d.GetBool("curated", false),
			},
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &MavenRemoteRepo{
			JavaRemoteRepo{
				RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
					Rclass:      rclass,
					PackageType: MavenPackageType,
				},
				SuppressPomConsistencyChecks: false,
			},
			RepositoryCurationParams{
				Curated: false,
			},
		}, nil
	}

	return mkResourceSchemaMaven(mavenRemoteSchema, packer.Default(mavenRemoteSchema), unpackMavenRemoteRepo, constructor)
}

var resourceMavenV1 = &schema.Resource{
	Schema: mavenRemoteSchemaV1,
}

// Old schema, the one needs to be migrated (seconds -> secs)
var mavenRemoteSchemaV1 = utilsdk.MergeMaps(
	JavaRemoteSchema(true, MavenPackageType, false),
	map[string]*schema.Schema{
		"metadata_retrieval_timeout_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      60,
			ValidateFunc: validation.IntAtLeast(0),
			Description:  "This value refers to the number of seconds to cache metadata files before checking for newer versions on remote server. A value of 0 indicates no caching. Cannot be larger than retrieval_cache_period_seconds attribute. Default value is 60.",
		},
	},
)

func mkResourceSchemaMaven(skeema map[string]*schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor repository.Constructor) *schema.Resource {
	var reader = repository.MkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: repository.MkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: repository.MkRepoUpdate(unpack, reader),
		DeleteContext: repository.DeleteRepo,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceMavenV1.CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceMavenStateUpgradeV1,
				Version: 1,
			},
		},

		Schema:        skeema,
		SchemaVersion: 2,
		CustomizeDiff: repository.ProjectEnvironmentsDiff,
	}
}

func ResourceMavenStateUpgradeV1(_ context.Context, rawState map[string]interface{}, _ interface{}) (map[string]interface{}, error) {
	if rawState["metadata_retrieval_timeout_seconds"] != nil {
		rawState["metadata_retrieval_timeout_secs"] = rawState["metadata_retrieval_timeout_seconds"]
		delete(rawState, "metadata_retrieval_timeout_seconds")
	}

	return rawState, nil
}
