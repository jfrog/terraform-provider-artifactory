package remote

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	"github.com/jfrog/terraform-provider-shared/util"
)

const MavenPackageType = "maven"

func ResourceArtifactoryRemoteMavenRepository() *schema.Resource {
	mavenRemoteSchema := JavaRemoteSchema(true, MavenPackageType, false)

	var unpackMavenRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		repo := UnpackJavaRemoteRepo(data, MavenPackageType)
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &JavaRemoteRepo{
			RepositoryRemoteBaseParams: RepositoryRemoteBaseParams{
				Rclass:      rclass,
				PackageType: MavenPackageType,
			},
			SuppressPomConsistencyChecks: false,
		}, nil
	}

	return mkResourceSchemaMaven(mavenRemoteSchema, packer.Default(mavenRemoteSchema), unpackMavenRemoteRepo, constructor)
}

var resourceMavenV1 = &schema.Resource{
	Schema: mavenRemoteSchemaV1,
}

// Old schema, the one needs to be migrated (seconds -> secs)
var mavenRemoteSchemaV1 = util.MergeMaps(
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
