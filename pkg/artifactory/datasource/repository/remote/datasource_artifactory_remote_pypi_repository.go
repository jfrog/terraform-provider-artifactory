package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/datasource/repository"
	resource_repository "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/samber/lo"
)

type PypiRemoteRepo struct {
	remote.RepositoryRemoteBaseParams
	remote.RepositoryCurationParams
	PypiRegistryUrl      string `json:"pyPIRegistryUrl"`
	PypiRepositorySuffix string `json:"pyPIRepositorySuffix"`
}

var PyPiSchema = lo.Assign(
	remote.BaseSchema,
	remote.CurationRemoteRepoSchema,
	map[string]*schema.Schema{
		"pypi_registry_url": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "https://pypi.org",
			ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
			Description:      "To configure the remote repo to proxy public external PyPI repository, or a PyPI repository hosted on another Artifactory server. See JFrog Pypi documentation for the usage details. Default value is 'https://pypi.org'.",
		},
		"pypi_repository_suffix": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "simple",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			Description:      "Usually should be left as a default for 'simple', unless the remote is a PyPI server that has custom registry suffix, like +simple in DevPI. Default value is 'simple'.",
		},
	},
	resource_repository.RepoLayoutRefSDKv2Schema(remote.Rclass, resource_repository.PyPiPackageType),
)

var PyPiSchemas = remote.GetSchemas(PyPiSchema)

func DataSourceArtifactoryRemotePypiRepository() *schema.Resource {
	constructor := func() (interface{}, error) {
		repoLayout, err := resource_repository.GetDefaultRepoLayoutRef(remote.Rclass, resource_repository.PyPiPackageType)
		if err != nil {
			return nil, err
		}

		return &PypiRemoteRepo{
			RepositoryRemoteBaseParams: remote.RepositoryRemoteBaseParams{
				Rclass:        remote.Rclass,
				PackageType:   resource_repository.PyPiPackageType,
				RepoLayoutRef: repoLayout,
			},
		}, nil
	}

	pypiSchema := getSchema(PyPiSchemas)

	return &schema.Resource{
		Schema:      pypiSchema,
		ReadContext: repository.MkRepoReadDataSource(packer.Default(pypiSchema), constructor),
		Description: "Provides a data source for a remote Pypi repository",
	}
}
