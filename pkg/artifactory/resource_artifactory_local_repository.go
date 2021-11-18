package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var legacyLocalSchema = mergeSchema(map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repoKeyValidator,
	},
	"package_type": {
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Computed:     true,
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
}, compressionFormats)

func resourceArtifactoryLocalRepository() *schema.Resource {
	return mkResourceSchema(legacyLocalSchema, inSchema(legacyLocalSchema), unmarshalLocalRepository, func() interface{} {
		return &MessyLocalRepo{
			LocalRepositoryBaseParams: LocalRepositoryBaseParams{
				Rclass: "local",
			},
		}
	})
}

type CommonMavenGradleLocalRepositoryParams struct {
	MaxUniqueSnapshots           int    `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots,omitempty"`
	HandleReleases               *bool  `hcl:"handle_releases" json:"handleReleases,omitempty"`
	HandleSnapshots              *bool  `hcl:"handle_snapshots" json:"handleSnapshots,omitempty"`
	SuppressPomConsistencyChecks *bool  `hcl:"suppress_pom_consistency_checks" json:"suppressPomConsistencyChecks,omitempty"`
	SnapshotVersionBehavior      string `hcl:"snapshot_version_behavior" json:"snapshotVersionBehavior,omitempty"`
	ChecksumPolicyType           string `hcl:"checksum_policy_type" json:"checksumPolicyType,omitempty"`
}

type MessyDebianLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	DebianTrivialLayout *bool `hcl:"debian_trivial_layout" json:"debianTrivialLayout,omitempty"`
}

type RpmLocalRepositoryParams struct {
	LocalRepositoryBaseParams
	YumRootDepth            int   `hcl:"yum_root_depth" json:"yumRootDepth,omitempty"`
	CalculateYumMetadata    *bool `hcl:"calculate_yum_metadata" json:"calculateYumMetadata,omitempty"`
	EnableFileListsIndexing *bool `hcl:"enable_file_lists_indexing" json:"enableFileListsIndexing,omitempty"`
}

type MessyLocalRepo struct {
	LocalRepositoryBaseParams
	CommonMavenGradleLocalRepositoryParams
	MessyDebianLocalRepositoryParams
	DockerLocalRepositoryParams
	RpmLocalRepositoryParams
	ForceNugetAuthentication bool `hcl:"force_nuget_authentication" json:"forceNugetAuthentication"`
}

func unmarshalLocalRepository(data *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{ResourceData: data}
	repo := MessyLocalRepo{}
	repo.Rclass = "local"
	repo.Key = d.getString("key", false)
	repo.PackageType = d.getString("package_type", false)
	repo.Description = d.getString("description", false)
	repo.Notes = d.getString("notes", false)
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.IncludesPattern = d.getString("includes_pattern", false)
	repo.ExcludesPattern = d.getString("excludes_pattern", false)
	repo.RepoLayoutRef = d.getString("repo_layout_ref", false)
	repo.MaxUniqueTags = d.getInt("max_unique_tags", false)
	repo.BlackedOut = d.getBoolRef("blacked_out", false)
	repo.CalculateYumMetadata = d.getBoolRef("calculate_yum_metadata", false)
	repo.YumRootDepth = d.getInt("yum_root_depth", false)
	repo.ArchiveBrowsingEnabled = d.getBoolRef("archive_browsing_enabled", false)
	repo.DockerApiVersion = d.getString("docker_api_verision", false)
	if repo.DockerApiVersion == "" {
		repo.DockerApiVersion = "V2" // for backward compatibility
	}
	repo.EnableFileListsIndexing = d.getBoolRef("enable_file_lists_indexing", false)
	repo.PropertySets = d.getSet("property_sets")
	repo.HandleReleases = d.getBoolRef("handle_releases", false)
	repo.HandleSnapshots = d.getBoolRef("handle_snapshots", false)
	repo.ChecksumPolicyType = d.getString("checksum_policy_type", false)
	repo.MaxUniqueSnapshots = d.getInt("max_unique_snapshots", false)
	repo.SnapshotVersionBehavior = d.getString("snapshot_version_behavior", false)
	repo.SuppressPomConsistencyChecks = d.getBoolRef("suppress_pom_consistency_checks", false)
	repo.XrayIndex = d.getBoolRef("xray_index", false)
	repo.ForceNugetAuthentication = d.getBool("force_nuget_authentication", false)

	return repo, repo.Key, nil
}
