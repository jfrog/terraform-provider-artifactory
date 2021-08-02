package artifactory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

func unpackVirtualRepository(s *schema.ResourceData) *v1.VirtualRepository {
	d := &ResourceData{s}
	repo := new(v1.VirtualRepository)

	repo.Key = d.getStringRef("key", false)
	repo.RClass = artifactory.String("virtual")
	repo.PackageType = d.getStringRef("package_type", false)
	repo.IncludesPattern = d.getStringRef("includes_pattern", false)
	repo.ExcludesPattern = d.getStringRef("excludes_pattern", false)
	repo.RepoLayoutRef = d.getStringRef("repo_layout_ref", false)
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.getBoolRef("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.Repositories = d.getListRef("repositories")
	repo.Description = d.getStringRef("description", false)
	repo.Notes = d.getStringRef("notes", false)
	repo.KeyPair = d.getStringRef("key_pair", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.getStringRef("pom_repository_references_cleanup_policy", false)
	repo.DefaultDeploymentRepo = d.getStringRef("default_deployment_repo", false)
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)

	return repo
}

func packVirtualRepository(repo *v1.VirtualRepository, d *schema.ResourceData) error {
	hasErr := false
	logErr := cascadingErr(&hasErr)

	logErr(d.Set("key", repo.Key))
	logErr(d.Set("package_type", repo.PackageType))
	logErr(d.Set("description", repo.Description))
	logErr(d.Set("notes", repo.Notes))
	logErr(d.Set("includes_pattern", repo.IncludesPattern))
	logErr(d.Set("excludes_pattern", repo.ExcludesPattern))
	logErr(d.Set("repo_layout_ref", repo.RepoLayoutRef))
	logErr(d.Set("debian_trivial_layout", repo.DebianTrivialLayout))
	logErr(d.Set("artifactory_requests_can_retrieve_remote_artifacts", repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts))
	logErr(d.Set("key_pair", repo.KeyPair))
	logErr(d.Set("pom_repository_references_cleanup_policy", repo.PomRepositoryReferencesCleanupPolicy))
	logErr(d.Set("default_deployment_repo", repo.DefaultDeploymentRepo))
	logErr(d.Set("repositories", repo.Repositories))
	logErr(d.Set("force_nuget_authentication", repo.ForceNugetAuthentication))

	if hasErr {
		return fmt.Errorf("failed to pack virtual repo")
	}

	return nil
}

func resourceVirtualRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repo := unpackVirtualRepository(d)

	_, err := c.V1.Repositories.CreateVirtual(context.Background(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceVirtualRepositoryRead(d, m)
}

func resourceVirtualRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repo, resp, err := c.V1.Repositories.GetVirtual(context.Background(), d.Id())

	if resp == nil {
		return fmt.Errorf("no response returned in resourceVirtualRepositoryRead")
	}

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	return packVirtualRepository(repo, d)
}

func resourceVirtualRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repo := unpackVirtualRepository(d)

	_, err := c.V1.Repositories.UpdateVirtual(context.Background(), d.Id(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceVirtualRepositoryRead(d, m)
}

func resourceVirtualRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld
	repo := unpackVirtualRepository(d)

	resp, err := c.V1.Repositories.DeleteVirtual(context.Background(), *repo.Key)

	if resp == nil {
		return fmt.Errorf("no response returned in resourceVirtualRepositoryDelete")
	}

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	}
	return err
}

func resourceVirtualRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*ArtClient).ArtOld

	key := d.Id()
	_, resp, err := c.V1.Repositories.GetVirtual(context.Background(), key)

	if resp == nil {
		// this really should be nil, err because we truly have no idea what the state is
		return false, fmt.Errorf("no response returned in resourceVirtualRepositoryExists")
	}

	// Cannot check for 404 because artifactory returns 400
	if resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}

	return true, err
}
