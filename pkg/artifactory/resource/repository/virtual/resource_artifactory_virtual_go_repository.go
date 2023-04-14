package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/util"
)

const GoPackageType = "go"

var GoVirtualSchema = util.MergeMaps(BaseVirtualRepoSchema, map[string]*schema.Schema{

	"external_dependencies_enabled": {
		Type:        schema.TypeBool,
		Default:     true,
		Optional:    true,
		Description: "When set (default), Artifactory will automatically follow remote VCS roots in 'go-import' meta tags to download remote modules.",
	},
	"external_dependencies_patterns": {
		Type:     schema.TypeList,
		Optional: true,
		ForceNew: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		RequiredWith: []string{"external_dependencies_enabled"},
		Description: "An allow list of Ant-style path patterns that determine which remote VCS roots Artifactory will " +
			"follow to download remote modules from, when presented with 'go-import' meta tags in the remote repository response.",
	},
}, repository.RepoLayoutRefSchema(Rclass, GoPackageType))

func ResourceArtifactoryVirtualGoRepository() *schema.Resource {
	type GoVirtualRepositoryParams struct {
		RepositoryBaseParams
		ExternalDependenciesEnabled  bool     `hcl:"external_dependencies_enabled" json:"externalDependenciesEnabled,omitempty"`
		ExternalDependenciesPatterns []string `hcl:"external_dependencies_patterns" json:"externalDependenciesPatterns,omitempty"`
	}

	var unpackGoVirtualRepository = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &util.ResourceData{ResourceData: s}

		repo := GoVirtualRepositoryParams{
			RepositoryBaseParams:         UnpackBaseVirtRepo(s, GoPackageType),
			ExternalDependenciesPatterns: d.GetList("external_dependencies_patterns"),
			ExternalDependenciesEnabled:  d.GetBool("external_dependencies_enabled", false),
		}
		repo.PackageType = "go"
		return &repo, repo.Key, nil
	}

	constructor := func() (interface{}, error) {
		return &GoVirtualRepositoryParams{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      Rclass,
				PackageType: GoPackageType,
			},
		}, nil
	}

	return repository.MkResourceSchema(
		GoVirtualSchema,
		packer.Default(GoVirtualSchema),
		unpackGoVirtualRepository,
		constructor,
	)
}
