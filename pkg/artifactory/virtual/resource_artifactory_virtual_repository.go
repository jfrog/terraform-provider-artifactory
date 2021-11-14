package virtual

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/repos"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/validators"
)

var legacyVirtualSchema = map[string]*schema.Schema{
	"key": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
	"package_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validators.RepoTypeValidator,
	},
	"repositories": {
		Type:     schema.TypeList,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Required: true,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"notes": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"includes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Default:  "**/*",
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"repo_layout_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"debian_trivial_layout": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"artifactory_requests_can_retrieve_remote_artifacts": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"key_pair": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"pom_repository_references_cleanup_policy": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"default_deployment_repo": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"force_nuget_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
}

func ResourceArtifactoryVirtualRepository() *schema.Resource {
	packer := util.UniversalPack(util.SchemaHasKey(legacyVirtualSchema))
	skeema := repos.MkResourceSchema(legacyVirtualSchema, packer, unpackVirtualRepository, func() interface{} {
		return &MessyVirtualRepo{
			RepositoryBaseParams: RepositoryBaseParams{
				Rclass: "virtual",
			},
		}
	})
	skeema.DeprecationMessage = "This resource is deprecated and you should use repo type specific resources " +
		"(such as artifactory_virtual_maven_repository) in the future"
	return skeema
}

type DebianVirtualRepositoryParams struct {
	RepositoryBaseParams
	DebianTrivialLayout *bool `json:"debianTrivialLayout,omitempty"`
}
type NugetVirtualRepositoryParams struct {
	RepositoryBaseParams
	ForceNugetAuthentication *bool `json:"forceNugetAuthentication,omitempty"`
}
type MessyVirtualRepo struct {
	RepositoryBaseParams
	DebianVirtualRepositoryParams
	MavenVirtualRepositoryParams
	NugetVirtualRepositoryParams
}

func unpackVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &util.ResourceData{s}
	repo := MessyVirtualRepo{}

	repo.Key = d.GetString("key", false)
	repo.Rclass = "virtual"
	repo.PackageType = d.GetString("package_type", false)
	repo.IncludesPattern = d.GetString("includes_pattern", false)
	repo.ExcludesPattern = d.GetString("excludes_pattern", false)
	repo.RepoLayoutRef = d.GetString("repo_layout_ref", false)
	repo.DebianTrivialLayout = d.GetBoolRef("debian_trivial_layout", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.GetBool("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.Repositories = d.GetList("repositories")
	repo.Description = d.GetString("description", false)
	repo.Notes = d.GetString("notes", false)
	repo.KeyPair = d.GetString("key_pair", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.GetString("pom_repository_references_cleanup_policy", false)
	repo.DefaultDeploymentRepo = d.GetString("default_deployment_repo", false)
	// because this doesn't apply to all repo types, RT isn't required to honor what you tell it.
	// So, saying the type is "maven" but then setting this to 'true' doesn't make sense, and RT doesn't seem to care what you tell it
	repo.ForceNugetAuthentication = d.GetBoolRef("force_nuget_authentication", false)
	return &repo, repo.Key, nil
}
