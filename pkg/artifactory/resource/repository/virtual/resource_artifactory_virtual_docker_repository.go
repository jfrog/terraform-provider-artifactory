package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryVirtualDockerRepository() *schema.Resource {

	const packageType = "docker"

	dockerVirtualSchema := util.MergeMaps(BaseVirtualRepoSchema, map[string]*schema.Schema{
		"resolve_docker_tags_by_timestamp": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When enabled, in cases where the same Docker tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp.",
		},
	}, repository.RepoLayoutRefSchema("virtual", packageType))

	type DockerVirtualRepositoryParams struct {
		RepositoryBaseParams
		ResolveDockerTagsByTimestamp bool `json:"resolveDockerTagsByTimestamp"`
	}

	unpackDockerVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := DockerVirtualRepositoryParams{
			RepositoryBaseParams:         UnpackBaseVirtRepo(data, "docker"),
			ResolveDockerTagsByTimestamp: d.GetBool("resolve_docker_tags_by_timestamp", false),
		}

		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DockerVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: packageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		dockerVirtualSchema,
		packer.Default(dockerVirtualSchema),
		unpackDockerVirtualRepository,
		constructor,
	)
}
