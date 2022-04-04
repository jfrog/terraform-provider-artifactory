package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		ValidateFunc: repoTypeValidator,
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

func resourceArtifactoryVirtualRepository() *schema.Resource {
	skeema := mkResourceSchema(legacyVirtualSchema, universalPack(legacyVirtualSchema), unpackVirtualRepository, func() interface{} {
		return &MessyVirtualRepo{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass: "virtual",
			},
		}
	})
	skeema.DeprecationMessage = "This resource is deprecated and you should use repo type specific resources " +
		"(such as artifactory_virtual_maven_repository) in the future"
	return skeema
}

type DebianVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	DebianTrivialLayout *bool `json:"debianTrivialLayout,omitempty"`
}

type NugetVirtualRepositoryParams struct {
	VirtualRepositoryBaseParams
	ForceNugetAuthentication *bool `json:"forceNugetAuthentication,omitempty"`
}

type MessyVirtualRepo struct {
	VirtualRepositoryBaseParams
	DebianVirtualRepositoryParams
	JavaVirtualRepositoryParams
	NugetVirtualRepositoryParams
}

func unpackVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := MessyVirtualRepo{}

	repo.Key = d.getString("key", false)
	repo.Rclass = "virtual"
	repo.PackageType = d.getString("package_type", false)
	repo.IncludesPattern = d.getString("includes_pattern", false)
	repo.ExcludesPattern = d.getString("excludes_pattern", false)
	repo.RepoLayoutRef = d.getString("repo_layout_ref", false)
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.getBool("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.Repositories = d.getList("repositories")
	repo.Description = d.getString("description", false)
	repo.Notes = d.getString("notes", false)
	repo.KeyPair = d.getString("key_pair", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.getString("pom_repository_references_cleanup_policy", false)
	repo.DefaultDeploymentRepo = handleResetWithNonExistantValue(d, "default_deployment_repo")
	// because this doesn't apply to all repo types, RT isn't required to honor what you tell it.
	// So, saying the type is "maven" but then setting this to 'true' doesn't make sense, and RT doesn't seem to care what you tell it
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)
	return &repo, repo.Key, nil
}
