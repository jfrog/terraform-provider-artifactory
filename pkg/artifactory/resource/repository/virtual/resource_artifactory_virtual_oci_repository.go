package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v10/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

const OciPackageType = "oci"

var OciVirtualSchema = utilsdk.MergeMaps(
	BaseVirtualRepoSchema,
	map[string]*schema.Schema{
		"resolve_oci_tags_by_timestamp": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "When enabled, in cases where the same OCI tag exists in two or more of the aggregated repositories, Artifactory will return the tag that has the latest timestamp.",
		},
	},
	repository.RepoLayoutRefSchema(Rclass, OciPackageType),
)

func ResourceArtifactoryVirtualOciRepository() *schema.Resource {
	type OciVirtualRepositoryParams struct {
		RepositoryBaseParams
		ResolveOciTagsByTimestamp bool `hcl:"resolve_oci_tags_by_timestamp" json:"resolveDockerTagsByTimestamp"`
	}

	unpackOciVirtualRepository := func(data *schema.ResourceData) (interface{}, string, error) {
		d := &utilsdk.ResourceData{ResourceData: data}
		repo := OciVirtualRepositoryParams{
			RepositoryBaseParams:      UnpackBaseVirtRepo(data, OciPackageType),
			ResolveOciTagsByTimestamp: d.GetBool("resolve_oci_tags_by_timestamp", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &OciVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: OciPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		OciVirtualSchema,
		packer.Default(OciVirtualSchema),
		unpackOciVirtualRepository,
		constructor,
	)
}
