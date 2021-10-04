package artifactory

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

var baseVirtualRepoSchema = map[string]*schema.Schema{
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
		Description: "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. " +
			"When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of x/y/**/z/*." +
			"By default no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"repositories": {
		Type:     schema.TypeList,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Required: true,
	},

	"artifactory_requests_can_retrieve_remote_artifacts": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"default_deployment_repo": {
		Type:     schema.TypeString,
		Optional: true,
	},
}
var legacySchema = mergeSchema(baseVirtualRepoSchema, map[string]*schema.Schema{

	"debian_trivial_layout": {
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

	"force_nuget_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
})

func newMessyRepo() interface{} {
	return &MessyVirtualRepo{}
}

var readFunc = mkRepoRead(packVirtualRepository, newMessyRepo)

func resourceArtifactoryVirtualRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkRepoCreate(unpackVirtualRepository, readFunc),
		Read:   readFunc,
		Update: mkRepoUpdate(unpackVirtualRepository, readFunc),
		Delete: deleteRepo,
		Exists: resourceVirtualRepositoryExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: legacySchema,
		DeprecationMessage: "This resource is deprecated and you should use repo type specific resources " +
			"(such as artifactory_virtual_maven_repository) in the future",
	}
}

type MessyVirtualRepo struct {
	services.VirtualRepositoryBaseParams
	services.DebianVirtualRepositoryParams
	services.MavenVirtualRepositoryParams
	services.NugetVirtualRepositoryParams
}

func unpackBaseVirtRepo(s *schema.ResourceData) (services.VirtualRepositoryBaseParams, string) {
	d := &ResourceData{s}

	repo := services.VirtualRepositoryBaseParams{}

	repo.Key = d.getString("key", false)
	repo.Rclass = "virtual"
	repo.PackageType = d.getString("package_type", false)
	repo.IncludesPattern = d.getString("includes_pattern", false)
	repo.ExcludesPattern = d.getString("excludes_pattern", false)
	repo.RepoLayoutRef = d.getString("repo_layout_ref", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.getBoolRef("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.Repositories = d.getList("repositories")
	repo.Description = d.getString("description", false)
	repo.Notes = d.getString("notes", false)
	repo.DefaultDeploymentRepo = d.getString("default_deployment_repo", false)

	return repo, repo.Key
}

func unpackVirtualRepository(s *schema.ResourceData) (interface{}, string) {
	d := &ResourceData{s}
	base, _ := unpackBaseVirtRepo(s)
	repo := MessyVirtualRepo{
		VirtualRepositoryBaseParams: base,
	}
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.getBoolRef("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.KeyPair = d.getString("key_pair", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.getString("pom_repository_references_cleanup_policy", false)
	// because this doesn't apply to all repo types, RT isn't required to honor what you tell it.
	// So, saying the type is "maven" but then setting this to 'true' doesn't make sense, and RT doesn't seem to care what you tell it
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)

	return &repo, repo.Key
}
func packBaseVirtRepo(d *schema.ResourceData, repo services.VirtualRepositoryBaseParams)  Lens {
	setValue := mkLens(d)

	setValue("key", repo.Key)
	setValue("package_type", repo.PackageType)
	setValue("description", repo.Description)
	setValue("notes", repo.Notes)
	setValue("includes_pattern", repo.IncludesPattern)
	setValue("excludes_pattern", repo.ExcludesPattern)
	setValue("repo_layout_ref", repo.RepoLayoutRef)
	setValue("artifactory_requests_can_retrieve_remote_artifacts", repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	setValue("default_deployment_repo", repo.DefaultDeploymentRepo)
	setValue("repositories", repo.Repositories)
	return setValue
}
func packVirtualRepository(r interface{}, d *schema.ResourceData) error {
	repo := r.(*MessyVirtualRepo)
	setValue := packBaseVirtRepo(d, repo.VirtualRepositoryBaseParams)

	setValue("debian_trivial_layout", repo.DebianTrivialLayout)
	setValue("artifactory_requests_can_retrieve_remote_artifacts", repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	setValue("key_pair", repo.KeyPair)
	setValue("pom_repository_references_cleanup_policy", repo.PomRepositoryReferencesCleanupPolicy)
	setValue("repositories", repo.Repositories)
	errors := setValue("force_nuget_authentication", repo.ForceNugetAuthentication)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack virtual repo %q", errors)
	}

	return nil
}






func resourceVirtualRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := m.(*resty.Client).R().Head(repositoriesEndpoint + d.Id())

	return err == nil, err
}
