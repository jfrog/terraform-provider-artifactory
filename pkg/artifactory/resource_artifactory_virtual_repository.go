package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
	skeema := mkResourceSchema(legacyVirtualSchema, inSchema(legacyVirtualSchema), unpackVirtualRepository, func() interface{} {
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

func resourceArtifactoryVirtualGenericRepository(pkt string) *schema.Resource {
	constructor := func() interface{} {
		return &VirtualRepositoryBaseParams{
			PackageType: pkt,
			Rclass:      "virtual",
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseVirtRepo(data, pkt)
		return repo, repo.Id(), nil
	}
	return mkResourceSchema(getBaseVirtualRepoSchema(pkt), defaultPacker, unpack, constructor)
}

func resourceArtifactoryVirtualRepositoryWithRetrievalCachePeriodSecs(pkt string) *schema.Resource {
	var repoWithRetrivalCachePeriodSecsVirtualSchema = mergeSchema(getBaseVirtualRepoSchema(pkt), map[string]*schema.Schema{
		"retrieval_cache_period_seconds": {
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      7200,
			Description:  "This value refers to the number of seconds to cache metadata files before checking for newer versions on aggregated repositories. A value of 0 indicates no caching.",
			ValidateFunc: validation.IntAtLeast(0),
		},
	})
	constructor := func() interface{} {
		return &VirtualRepositoryBaseParamsWithRetrievalCachePeriodSecs{
			VirtualRepositoryBaseParams: VirtualRepositoryBaseParams{
				Rclass:      "virtual",
				PackageType: pkt,
			},
		}
	}
	unpack := func(data *schema.ResourceData) (interface{}, string, error) {
		repo := unpackBaseVirtRepoWithRetrievalCachePeriodSecs(data, pkt)
		return repo, repo.Id(), nil
	}
	return mkResourceSchema(repoWithRetrivalCachePeriodSecsVirtualSchema, defaultPacker, unpack, constructor)
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
	MavenVirtualRepositoryParams
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
