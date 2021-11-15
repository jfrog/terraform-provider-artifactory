package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
)

var cargoRemoteSchema = util.MergeSchema(baseRemoteSchema, map[string]*schema.Schema{
	"git_registry_url": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.IsURLWithHTTPorHTTPS,
		Description:  `This is the index url, expected to be a git repository. for remote artifactory use "arturl/git/repokey.git"`,
	},
	"anonymous_access": {
		Type:     schema.TypeBool,
		Optional: true,
		Description: "(On the UI: Anonymous download and search) Cargo client does not send credentials when performing download and search for crates. " +
			"Enable this to allow anonymous access to these resources (only), note that this will override the security anonymous access option.",
	},
})

type CargoRemoteRepo struct {
	RepositoryBaseParams
	RegistryUrl     string `hcl:"git_registry_url" json:"gitRegistryUrl"`
	AnonymousAccess bool   `hcl:"anonymous_access" json:"cargoAnonymousAccess"`
}

func ResourceArtifactoryRemoteCargoRepository() *schema.Resource {
	return repos.MkResourceSchema(cargoRemoteSchema, util.DefaultPacker, unpackCargoRemoteRepo, func() interface{} {
		return &CargoRemoteRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass:      "remote",
				PackageType: "cargo",
			},
		}
	})
}

func unpackCargoRemoteRepo(s *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{ResourceData: s}
	repo := CargoRemoteRepo{
		RepositoryBaseParams: unpackBaseRemoteRepo(s, "cargo"),
		RegistryUrl:          d.GetString("git_registry_url", false),
		AnonymousAccess:      d.GetBool("anonymous_access", false),
	}
	return repo, repo.Id(), nil
}
