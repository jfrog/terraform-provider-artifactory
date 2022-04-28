package federated

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/validator"
)

var FederatedRepoTypesSupported = []string{
	"alpine",
	"bower",
	"cargo",
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"debian",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"gradle",
	"helm",
	"ivy",
	"maven",
	"npm",
	"nuget",
	"opkg",
	"puppet",
	"pypi",
	"rpm",
	"sbt",
	"vagrant",
}

var BaseFederatedRepoSchema = map[string]*schema.Schema{
	"key": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repository.RepoKeyValidator,
		Description: "A mandatory identifier for the repository that must be unique. It cannot begin with a number or" +
			" contain spaces or special characters",
	},
	"project_key": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validator.ProjectKey,
		Description:      "Project key for assigning this repository to. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		MinItems:    1,
		MaxItems:    2,
		Set:         schema.HashString,
		Optional:    true,
		Description: `Project environment for assigning this repository to. Allow values: "DEV" or "PROD"`,
	},
	"package_type": {
		Type:     schema.TypeString,
		Required: false,
		Computed: true,
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
		Computed: true,
		Description: "List of artifact patterns to include when evaluating artifact requests in the form of x/y/**/z/*. " +
			"When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (**/*).",
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of" +
			" x/y/**/z/*. By default, no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: repository.ValidateRepoLayoutRefSchemaOverride,
		Description: "Sets the layout that the repository should use for storing and identifying modules. " +
			"A recommended layout that corresponds to the package type defined is suggested, and index packages " +
			"uploaded and calculate metadata accordingly.",
	},
	"blacked_out": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, the repository does not participate in artifact resolution and new artifacts cannot be deployed.",
	},
	"xray_index": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
		Description: "Enable Indexing In Xray. Repository will be indexed with the default retention period. " +
			"You will be able to change it via Xray settings.",
	},
	"priority_resolution": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
	},
	"property_sets": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Set:         schema.HashString,
		Optional:    true,
		Description: "List of property set names.",
	},
	"archive_browsing_enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Description: "When set, you may view content such as HTML or Javadoc files directly from Artifactory." +
			"This may not be safe and therefore requires strict content moderation to prevent malicious users from " +
			"uploading content that may compromise security (e.g., cross-site scripting attacks).",
	},
	"download_direct": {
		Type:     schema.TypeBool,
		Optional: true,
		Description: "When set, download requests to this repository will redirect the client to download " +
			"the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.",
	},
}
