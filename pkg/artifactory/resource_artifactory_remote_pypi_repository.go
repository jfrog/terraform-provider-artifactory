package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var pypiRemoteSchema = mergeSchema(baseRemoteSchema, map[string]*schema.Schema{
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
	"list_remote_folder_items": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: `(Optional) Lists the items of remote folders in simple and list browsing. The remote content is cached according to the value of the 'Retrieval Cache Period'. Default value is 'false'.`,
	},
})

type PypiRemoteRepo struct {
	RemoteRepositoryBaseParams
	PypiRegistryUrl       string `json:"pyPIRegistryUrl"`
	PypiRepositorySuffix  string `json:"pyPIRepositorySuffix"`
	ListRemoteFolderItems bool   `json:"listRemoteFolderItems"`
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
		PypiRegistryUrl:            d.getString("pypi_registry_url", false),
		PypiRepositorySuffix:       d.getString("pypi_repository_suffix", false),
		ListRemoteFolderItems:      d.getBool("list_remote_folder_items", false),
	}
	return repo, repo.Id(), nil
}
