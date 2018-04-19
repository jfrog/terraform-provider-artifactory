package artifactory

import (
	"context"
	"github.com/atlassian/go-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
	"net/http"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Default:  "discard_active_reference",
			},
			"default_deployment_repo": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func unmarshalVirtualRepository(s *schema.ResourceData) *artifactory.VirtualRepository {
	d := &ResourceData{s}
	repo := new(artifactory.VirtualRepository)

	repo.Key = d.GetStringRef("key")
	repo.RClass = artifactory.String("virtual")
	repo.PackageType = d.GetStringRef("package_type")
	repo.IncludesPattern = d.GetStringRef("includes_pattern")
	repo.ExcludesPattern = d.GetStringRef("excludes_pattern")
	repo.DebianTrivialLayout = d.GetBoolRef("debian_trivial_layout")
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.GetBoolRef("artifactory_requests_can_retrieve_remote_artifacts")
	repo.Repositories = d.GetListRef("repositories")
	repo.Description = d.GetStringRef("description")
	repo.Notes = d.GetStringRef("notes")
	repo.KeyPair = d.GetStringRef("key_pair")
	repo.PomRepositoryReferencesCleanupPolicy = artifactory.String(d.Get("pom_repository_references_cleanup_policy").(string))
	repo.DefaultDeploymentRepo = d.GetStringRef("default_deployment_repo")

	return repo
}

func marshalVirtualRepository(repo *artifactory.VirtualRepository, d *schema.ResourceData) {
	d.Set("key", repo.Key)
	d.Set("type", repo.RClass)
	d.Set("package_type", repo.PackageType)
	d.Set("description", repo.Description)
	d.Set("notes", repo.Notes)
	d.Set("includes_pattern", repo.IncludesPattern)
	d.Set("excludes_pattern", repo.ExcludesPattern)
	d.Set("debian_trivial_layout", repo.DebianTrivialLayout)
	d.Set("artifactory_requests_can_retrieve_remote_artifacts", repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	d.Set("key_pair", repo.KeyPair)
	d.Set("pom_repository_references_cleanup_policy", repo.PomRepositoryReferencesCleanupPolicy)
	d.Set("default_deployment_repo", repo.DefaultDeploymentRepo)
	d.Set("repositories", repo.Repositories)

}

func resourceVirtualRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	repo := unmarshalVirtualRepository(d)

	_, err := c.Repositories.CreateVirtual(context.Background(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceVirtualRepositoryRead(d, m)
}

func resourceVirtualRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	repo, resp, err := c.Repositories.GetVirtual(context.Background(), d.Id())
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	marshalVirtualRepository(repo, d)
	return nil
}

func resourceVirtualRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)

	repo := unmarshalVirtualRepository(d)

	_, err := c.Repositories.UpdateVirtual(context.Background(), d.Id(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceVirtualRepositoryRead(d, m)
}

func resourceVirtualRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Client)
	repo := unmarshalVirtualRepository(d)

	resp, err := c.Repositories.DeleteVirtual(context.Background(), *repo.Key)
	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}
	return err
}

func resourceVirtualRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*artifactory.Client)

	key := d.Id()
	_, resp, err := c.Repositories.GetVirtual(context.Background(), key)

	// Cannot check for 404 because artifactory returns 400
	if resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}

	return true, err
}
