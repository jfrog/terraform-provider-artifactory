package artifactory

import (
	"fmt"
	"net/http"

	"github.com/jfrog/jfrog-client-go/artifactory/services"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceArtifactoryVirtualRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualRepositoryCreate,
		Read:   resourceVirtualRepositoryRead,
		Update: resourceVirtualRepositoryUpdate,
		Delete: resourceVirtualRepositoryDelete,
		Exists: resourceVirtualRepositoryExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
		},
	}
}

type MessyVirtualRepo struct {
	services.VirtualRepositoryBaseParams
	services.DebianVirtualRepositoryParams
	services.MavenVirtualRepositoryParams
	services.NugetVirtualRepositoryParams
}

func (repo MessyVirtualRepo) Id() string {
	return repo.Key
}
func unpackVirtualRepository(s *schema.ResourceData) MessyVirtualRepo {
	d := &ResourceData{s}
	repo := MessyVirtualRepo{}

	repo.Key = d.getString("key", false)
	repo.Rclass = "virtual"
	repo.PackageType = d.getString("package_type", false)
	repo.IncludesPattern = d.getString("includes_pattern", false)
	repo.ExcludesPattern = d.getString("excludes_pattern", false)
	repo.RepoLayoutRef = d.getString("repo_layout_ref", false)
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.getBoolRef("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.Repositories = d.getList("repositories")
	repo.Description = d.getString("description", false)
	repo.Notes = d.getString("notes", false)
	repo.KeyPair = d.getString("key_pair", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.getString("pom_repository_references_cleanup_policy", false)
	repo.DefaultDeploymentRepo = d.getString("default_deployment_repo", false)
	// because this doesn't apply to all repo types, RT isn't required to honor what you tell it.
	// So, saying the type is "maven" but then setting this to 'true' doesn't make sense, and RT doesn't seem to care what you tell it
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)

	return repo
}

func packVirtualRepository(repo MessyVirtualRepo, d *schema.ResourceData) error {
	setValue := mkLens(d)

	setValue("key", repo.Key)
	setValue("package_type", repo.PackageType)
	setValue("description", repo.Description)
	setValue("notes", repo.Notes)
	setValue("includes_pattern", repo.IncludesPattern)
	setValue("excludes_pattern", repo.ExcludesPattern)
	setValue("repo_layout_ref", repo.RepoLayoutRef)
	setValue("debian_trivial_layout", repo.DebianTrivialLayout)
	setValue("artifactory_requests_can_retrieve_remote_artifacts", repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	setValue("key_pair", repo.KeyPair)
	setValue("pom_repository_references_cleanup_policy", repo.PomRepositoryReferencesCleanupPolicy)
	setValue("default_deployment_repo", repo.DefaultDeploymentRepo)
	setValue("repositories", repo.Repositories)
	errors := setValue("force_nuget_authentication", repo.ForceNugetAuthentication)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack virtual repo %q", errors)
	}

	return nil
}

func resourceVirtualRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty
	repo := unpackVirtualRepository(d)

	_, err := client.R().SetBody(repo).Put(repositoriesEndpoint + repo.Key)

	if err != nil {
		return err
	}
	d.SetId(repo.Key)
	return resourceVirtualRepositoryRead(d, m)
}

func resourceVirtualRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Resty
	repo := MessyVirtualRepo{}
	resp, err := c.R().SetResult(&repo).Get(repositoriesEndpoint + d.Id())

	if err != nil {
		if resp != nil && (resp.StatusCode() == http.StatusNotFound) {
			d.SetId("")
			return nil
		}
		return err
	}
	return packVirtualRepository(repo, d)
}

func resourceVirtualRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).Resty

	repo := unpackVirtualRepository(d)

	_, err := c.R().SetBody(repo).Post(repositoriesEndpoint + d.Id())
	if err != nil {
		return err
	}

	d.SetId(repo.Key)
	return resourceVirtualRepositoryRead(d, m)
}

func resourceVirtualRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).Resty

	resp, err := client.R().Delete(repositoriesEndpoint + d.Id())

	if err != nil && (resp != nil && resp.StatusCode() == http.StatusNotFound) {
		d.SetId("")
		return nil
	}
	return err
}

func resourceVirtualRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := m.(*ArtClient).Resty.R().Head(repositoriesEndpoint + d.Id())

	return err == nil, err
}
