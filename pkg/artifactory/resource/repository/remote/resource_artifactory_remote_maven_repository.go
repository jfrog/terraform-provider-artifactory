package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

func ResourceArtifactoryRemoteMavenRepository() *schema.Resource {
	mavenRemoteSchema := util.MergeMaps(
		getJavaRemoteSchema("maven", false),
		map[string]*schema.Schema{
			"metadata_retrieval_timeout_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				DefaultFunc: func() (interface{}, error) {
					return 60, nil
				},
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "This value refers to the number of seconds to cache metadata files before checking for newer versions on remote server. A value of 0 indicates no caching. Cannot be larger than retrieval_cache_period_seconds attribute. Default value is 60.",
			},
		},
	)

	type MavenRemoteRepo struct {
		JavaRemoteRepo
		MetadataRetrievalTimeoutSecs int `hcl:"metadata_retrieval_timeout_seconds" json:"metadataRetrievalTimeoutSecs"`
	}

	var unpackMavenRemoteRepo = func(data *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: data}
		repo := MavenRemoteRepo{
			JavaRemoteRepo:               UnpackJavaRemoteRepo(data, "maven"),
			MetadataRetrievalTimeoutSecs: d.GetInt("metadata_retrieval_timeout_seconds", false),
		}
		return repo, repo.Id(), nil
	}

	constructor := func() (interface{}, error) {
		return &MavenRemoteRepo{
			JavaRemoteRepo: JavaRemoteRepo{
				RepositoryBaseParams: RepositoryBaseParams{
					Rclass:      "remote",
					PackageType: "maven",
				},
				SuppressPomConsistencyChecks: false,
			},
		}, nil
	}

	return repository.MkResourceSchema(mavenRemoteSchema, packer.Default(mavenRemoteSchema), unpackMavenRemoteRepo, constructor)
}
