package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var pypiRemoteSchema = mergeSchema(baseRemoteSchema, map[string]*schema.Schema{
	"pypi_registry_url": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "https://pypi.org",
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
		Description:  `This is the index url, expected to be a git repository. for remote artifactory use "arturl/git/repokey.git"`,
	},
	"pypi_repository_suffix": {
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "simple",
		Description: "Usually should be left as a default for 'simple', unless the remote is a PyPI server that has custom registry suffix, like +simple in DevPI",
	},
})

type PypiRemoteRepo struct {
	RemoteRepositoryBaseParams
	RegistryUrl      string `hcl:"pypi_registry_url" json:"pyPIRegistryUrl"`
	RepositorySuffix string `hcl:"pypi_repository_suffix" json:"pyPIRepositorySuffix"`
}

func resourceArtifactoryRemotePypiRepository() *schema.Resource {
	return mkResourceSchema(pypiRemoteSchema, defaultPacker, unpackPypiRemoteRepo, func() interface{} {
		return &PypiRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:      "remote",
				PackageType: "pypi",
			},
		}
	})
}

func unpackPypiRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := PypiRemoteRepo{
		RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, "pypi"),
		RegistryUrl:                d.getString("pypi_registry_url", false),
		RepositorySuffix:           d.getString("pypi_repository_suffix", false),
	}
	return repo, repo.Id(), nil
}
