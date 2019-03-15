package artifactory

import (
	"context"
	"fmt"
	"net/http"

	"github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/atlassian/go-artifactory/v2/artifactory/v1"
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
				Default:  "generic",
				ForceNew: true,
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
				Default:  "",
			},
			"repo_layout_ref": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"handle_releases": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"handle_snapshots": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"max_unique_snapshots": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"debian_trivial_layout": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"checksum_policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "client-checksums",
			},
			"max_unique_tags": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"snapshot_version_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "non-unique",
			},
			"suppress_pom_consistency_checks": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"blacked_out": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
				Default:  "V2",
			},
			"enable_file_lists_indexing": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"xray_index": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func unmarshalLocalRepository(s *schema.ResourceData) *v1.LocalRepository {
	d := &ResourceData{s}

	repo := new(v1.LocalRepository)

	repo.RClass = artifactory.String("local")

	repo.Key = d.getStringRef("key")
	repo.PackageType = d.getStringRef("package_type")
	repo.Description = d.getStringRef("description")
	repo.Notes = d.getStringRef("notes")
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout")
	repo.IncludesPattern = d.getStringRef("includes_pattern")
	repo.ExcludesPattern = d.getStringRef("excludes_pattern")
	repo.RepoLayoutRef = d.getStringRef("repo_layout_ref")
	repo.MaxUniqueTags = d.getIntRef("max_unique_tags")
	repo.BlackedOut = d.getBoolRef("blacked_out")
	repo.CalculateYumMetadata = d.getBoolRef("calculate_yum_metadata")
	repo.YumRootDepth = d.getIntRef("yum_root_depth")
	repo.ArchiveBrowsingEnabled = d.getBoolRef("archive_browsing_enabled")
	repo.DockerApiVersion = d.getStringRef("docker_api_verision")
	repo.EnableFileListsIndexing = d.getBoolRef("enable_file_lists_indexing")
	repo.PropertySets = d.getSetRef("property_sets")
	repo.HandleReleases = d.getBoolRef("handle_releases")
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots")
	repo.ChecksumPolicyType = d.getStringRef("checksum_policy_type")
	repo.MaxUniqueSnapshots = d.getIntRef("max_unique_snapshots")
	repo.SnapshotVersionBehavior = d.getStringRef("snapshot_version_behavior")
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks")
	repo.XrayIndex = d.getBoolRef("xray_index")

	return repo
}

func resourceLocalRepositoryCreate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	repo := unmarshalLocalRepository(d)

	_, err := c.V1.Repositories.CreateLocal(context.Background(), repo)
	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceLocalRepositoryRead(d, m)
}

func resourceLocalRepositoryRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

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

		if hasErr {
			return fmt.Errorf("failed to marshal group")
		}
	}

	return err
}

func resourceLocalRepositoryUpdate(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)

	repo := unmarshalLocalRepository(d)
	_, err := c.V1.Repositories.UpdateLocal(context.Background(), d.Id(), repo)

	if err != nil {
		return err
	}

	d.SetId(*repo.Key)
	return resourceLocalRepositoryRead(d, m)
}

func resourceLocalRepositoryDelete(d *schema.ResourceData, m interface{}) error {
	c := m.(*artifactory.Artifactory)
	repo := unmarshalLocalRepository(d)

	resp, err := c.V1.Repositories.DeleteLocal(context.Background(), *repo.Key)

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}

func resourceLocalRepositoryExists(d *schema.ResourceData, m interface{}) (bool, error) {
	c := m.(*artifactory.Artifactory)

	_, resp, err := c.V1.Repositories.GetLocal(context.Background(), d.Id())

	// Cannot check for 404 because artifactory returns 400
	if resp.StatusCode == http.StatusBadRequest {
		return false, nil
	}

	return true, err
}
