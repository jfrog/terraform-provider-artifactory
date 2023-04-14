package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const DockerPackageType = "docker"

var DockerVirtualSchema = util.MergeMaps(BaseVirtualRepoSchema, map[string]*schema.Schema{
	"resolve_docker_tags_by_timestamp": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When enabled, in cases where the same Docker tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp.",
	},
}, repository.RepoLayoutRefSchema(Rclass, DockerPackageType))

func ResourceArtifactoryVirtualDockerRepository() *schema.Resource {

	type DockerVirtualRepositoryParams struct {
		RepositoryBaseParams
		ResolveDockerTagsByTimestamp bool `json:"resolveDockerTagsByTimestamp"`
	}

	unpackDockerVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := DockerVirtualRepositoryParams{
			RepositoryBaseParams:         UnpackBaseVirtRepo(data, DockerPackageType),
			ResolveDockerTagsByTimestamp: d.GetBool("resolve_docker_tags_by_timestamp", false),
		}

		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &DockerVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: DockerPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		DockerVirtualSchema,
		packer.Default(DockerVirtualSchema),
		unpackDockerVirtualRepository,
		constructor,
	)
}
