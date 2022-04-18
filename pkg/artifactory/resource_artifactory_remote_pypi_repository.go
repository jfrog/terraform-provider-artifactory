package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func resourceArtifactoryRemotePypiRepository() *schema.Resource {
	const packageType = "pypi"

	var pypiRemoteSchema = utils.MergeSchema(baseRemoteRepoSchema, map[string]*schema.Schema{
		"pypi_registry_url": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "https://pypi.org",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
			Description:      `(Optional) To configure the remote repo to proxy public external PyPI repository, or a PyPI repository hosted on another Artifactory server. See JFrog Pypi documentation for the usage details. Default value is 'https://pypi.org'.`,
		},
		"pypi_repository_suffix": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "simple",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      `(Optional) Usually should be left as a default for 'simple', unless the remote is a PyPI server that has custom registry suffix, like +simple in DevPI. Default value is 'simple'.`,
		},
	}, repoLayoutRefSchema("remote", packageType))

	type PypiRemoteRepo struct {
		RemoteRepositoryBaseParams
		PypiRegistryUrl      string `json:"pyPIRegistryUrl"`
		PypiRepositorySuffix string `json:"pyPIRepositorySuffix"`
	}

	var unpackPypiRemoteRepo = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &utils.ResourceData{s}
		repo := PypiRemoteRepo{
			RemoteRepositoryBaseParams: unpackBaseRemoteRepo(s, packageType),
			PypiRegistryUrl:            d.GetString("pypi_registry_url", false),
			PypiRepositorySuffix:       d.GetString("pypi_repository_suffix", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(pypiRemoteSchema, defaultPacker(pypiRemoteSchema), unpackPypiRemoteRepo, func() interface{} {
		return &PypiRemoteRepo{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:      "remote",
				PackageType: packageType,
			},
		}
	})
}
