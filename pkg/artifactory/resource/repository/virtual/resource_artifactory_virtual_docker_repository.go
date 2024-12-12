package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

var dockerSchema = lo.Assign(
	map[string]*schema.Schema{
		"resolve_docker_tags_by_timestamp": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When enabled, in cases where the same Docker tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp.",
		},
	},
	repository.RepoLayoutRefSDKv2Schema(Rclass, repository.DockerPackageType),
)

var DockerSchemas = GetSchemas(dockerSchema)

func ResourceArtifactoryVirtualDockerRepository() *schema.Resource {

	type DockerVirtualRepositoryParams struct {
		RepositoryBaseParams
		ResolveDockerTagsByTimestamp bool `json:"resolveDockerTagsByTimestamp"`
	}

	unpackDockerVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := DockerVirtualRepositoryParams{
			RepositoryBaseParams:         UnpackBaseVirtRepo(data, repository.DockerPackageType),
			ResolveDockerTagsByTimestamp: d.GetBool("resolve_docker_tags_by_timestamp", false),
		}

		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DockerVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: repository.DockerPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		DockerSchemas,
		packer.Default(DockerSchemas[CurrentSchemaVersion]),
		unpackDockerVirtualRepository,
		constructor,
	)
}
