package artifactory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	v1 "github.com/atlassian/go-artifactory/v2/artifactory/v1"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceArtifactoryLocalRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocalRepositoryCreate,
		Read:   resourceLocalRepositoryRead,
		Update: resourceLocalRepositoryUpdate,
		Delete: resourceLocalRepositoryDelete,
		Exists: resourceLocalRepositoryExists,
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
				Optional: true,
				ForceNew: true,
				Computed: true,
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
				Computed: true,
			},
			"excludes_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"repo_layout_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"handle_releases": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"handle_snapshots": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"max_unique_snapshots": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"debian_trivial_layout": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"checksum_policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"max_unique_tags": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"snapshot_version_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"suppress_pom_consistency_checks": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"blacked_out": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"property_sets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"archive_browsing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"calculate_yum_metadata": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"yum_root_depth": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"docker_api_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"enable_file_lists_indexing": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"xray_index": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"force_nuget_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func unmarshalLocalRepository(s *schema.ResourceData) *v1.LocalRepository {
	d := &ResourceData{s}

	repo := new(v1.LocalRepository)

	repo.RClass = artifactory.String("local")

	repo.Key = d.getStringRef("key", false)
	repo.PackageType = d.getStringRef("package_type", false)
	repo.Description = d.getStringRef("description", false)
	repo.Notes = d.getStringRef("notes", false)
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.IncludesPattern = d.getStringRef("includes_pattern", false)
	repo.ExcludesPattern = d.getStringRef("excludes_pattern", false)
	repo.RepoLayoutRef = d.getStringRef("repo_layout_ref", false)
	repo.MaxUniqueTags = d.getIntRef("max_unique_tags", false)
	repo.BlackedOut = d.getBoolRef("blacked_out", false)
	repo.CalculateYumMetadata = d.getBoolRef("calculate_yum_metadata", false)
	repo.YumRootDepth = d.getIntRef("yum_root_depth", false)
	repo.ArchiveBrowsingEnabled = d.getBoolRef("archive_browsing_enabled", false)
	repo.DockerApiVersion = d.getStringRef("docker_api_verision", false)
	repo.EnableFileListsIndexing = d.getBoolRef("enable_file_lists_indexing", false)
	repo.PropertySets = d.getSetRef("property_sets")
	repo.HandleReleases = d.getBoolRef("handle_releases", false)
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots", false)
	repo.ChecksumPolicyType = d.getStringRef("checksum_policy_type", false)
	repo.MaxUniqueSnapshots = d.getIntRef("max_unique_snapshots", false)
	repo.SnapshotVersionBehavior = d.getStringRef("snapshot_version_behavior", false)
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks", false)
	repo.XrayIndex = d.getBoolRef("xray_index", false)
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)

	return repo
}

func resourceLocalRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repo := unmarshalLocalRepository(d)

	_, err := c.V1.Repositories.CreateLocal(context.Background(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceLocalRepositoryRead(d, m)
}

func resourceLocalRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repo, resp, err := c.V1.Repositories.GetLocal(context.Background(), d.Id())

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
	} else if err == nil {
		hasErr := false
		logError := cascadingErr(&hasErr)

		logError(d.Set("key", repo.Key))
		logError(d.Set("package_type", repo.PackageType))
		logError(d.Set("description", repo.Description))
		logError(d.Set("notes", repo.Notes))
		logError(d.Set("includes_pattern", repo.IncludesPattern))
		logError(d.Set("excludes_pattern", repo.ExcludesPattern))
		logError(d.Set("repo_layout_ref", repo.RepoLayoutRef))
		logError(d.Set("debian_trivial_layout", repo.DebianTrivialLayout))
		logError(d.Set("max_unique_tags", repo.MaxUniqueTags))
		logError(d.Set("blacked_out", repo.BlackedOut))
		logError(d.Set("archive_browsing_enabled", repo.ArchiveBrowsingEnabled))
		logError(d.Set("calculate_yum_metadata", repo.CalculateYumMetadata))
		logError(d.Set("yum_root_depth", repo.YumRootDepth))
		logError(d.Set("docker_api_version", repo.DockerApiVersion))
		logError(d.Set("enable_file_lists_indexing", repo.EnableFileListsIndexing))
		logError(d.Set("property_sets", schema.NewSet(schema.HashString, castToInterfaceArr(*repo.PropertySets))))
		logError(d.Set("handle_releases", repo.HandleReleases))
		logError(d.Set("handle_snapshots", repo.HandleSnapshots))
		logError(d.Set("checksum_policy_type", repo.ChecksumPolicyType))
		logError(d.Set("max_unique_snapshots", repo.MaxUniqueSnapshots))
		logError(d.Set("snapshot_version_behavior", repo.SnapshotVersionBehavior))
		logError(d.Set("suppress_pom_consistency_checks", repo.SuppressPomConsistencyChecks))
		logError(d.Set("xray_index", repo.XrayIndex))
		logError(d.Set("force_nuget_authentication", repo.ForceNugetAuthentication))

		if hasErr {
			return fmt.Errorf("failed to marshal group")
		}
	}

	return err
}

func resourceLocalRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld

	repo := unmarshalLocalRepository(d)
	_, err := c.V1.Repositories.UpdateLocal(context.Background(), d.Id(), repo)

	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceLocalRepositoryRead(d, m)
}

func resourceLocalRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*ArtClient).ArtOld
	repo := unmarshalLocalRepository(d)

	resp, err := c.V1.Repositories.DeleteLocal(context.Background(), *repo.Key)

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}

func resourceLocalRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*ArtClient).ArtOld

	_, resp, err := c.V1.Repositories.GetLocal(context.Background(), d.Id())

	// Cannot check for 404 because artifactory returns 400
	if resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}

	return true, err
}
