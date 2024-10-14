package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/util"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"golang.org/x/exp/slices"

	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	"github.com/jfrog/terraform-provider-shared/validator"
)

const (
	AlpinePackageType            = "alpine"
	AnsiblePackageType           = "ansible"
	BowerPackageType             = "bower"
	CargoPackageType             = "cargo"
	ChefPackageType              = "chef"
	CocoapodsPackageType         = "cocoapods"
	ComposerPackageType          = "composer"
	CondaPackageType             = "conda"
	ConanPackageType             = "conan"
	CranPackageType              = "cran"
	DebianPackageType            = "debian"
	DockerPackageType            = "docker"
	GemsPackageType              = "gems"
	GenericPackageType           = "generic"
	GitLFSPackageType            = "gitlfs"
	GoPackageType                = "go"
	GradlePackageType            = "gradle"
	HelmPackageType              = "helm"
	HelmOCIPackageType           = "helmoci"
	HuggingFacePackageType       = "huggingfaceml"
	IvyPackageType               = "ivy"
	MavenPackageType             = "maven"
	NPMPackageType               = "npm"
	NugetPackageType             = "nuget"
	OCIPackageType               = "oci"
	OpkgPackageType              = "opkg"
	P2PackageType                = "p2"
	PubPackageType               = "pub"
	PuppetPackageType            = "puppet"
	PyPiPackageType              = "pypi"
	RPMPackageType               = "rpm"
	SBTPackageType               = "sbt"
	SwiftPackageType             = "swift"
	TerraformBackendPackageType  = "terraformbackend"
	TerraformModulePackageType   = "terraform_module"
	TerraformProviderPackageType = "terraform_provider"
	TerraformPackageType         = "terraform"
	VagrantPackageType           = "vagrant"
	VCSPackageType               = "vcs"
)

var BaseSchemaV1 = map[string]*schema.Schema{
	"key": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		ValidateDiagFunc: validator.RepoKey,
		Description:      "A mandatory identifier for the repository that must be unique. Must be 1 - 64 alphanumeric and hyphen characters. It cannot contain spaces or special characters.",
	},
	"project_key": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validator.ProjectKey,
		Description:      "Project key for assigning this repository to. Must be 2 - 32 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		MinItems: 0,
		MaxItems: 2,
		Set:      schema.HashString,
		Optional: true,
		Computed: true,
		Description: "Project environment for assigning this repository to. Allow values: \"DEV\", \"PROD\", or one of custom environment. " +
			"Before Artifactory 7.53.1, up to 2 values (\"DEV\" and \"PROD\") are allowed. From 7.53.1 onward, only one value is allowed. " +
			"The attribute should only be used if the repository is already assigned to the existing project. If not, " +
			"the attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create " +
			"state drift during the update.",
	},
	"package_type": {
		Type:     schema.TypeString,
		Required: false,
		Computed: true,
		ForceNew: true,
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Public description.",
	},
	"notes": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Internal description.",
	},
	"includes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Default:  "**/*",
		Description: "List of comma-separated artifact patterns to include when evaluating artifact requests in the form of `x/y/**/z/*`. " +
			"When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (`**/*`).",
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of `x/y/**/z/*`." +
			"By default no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:     schema.TypeString,
		Optional: true,
		// The default value in the UI is simple-default, in API maven-2-default. Provider will always override it ro math the UI.
		ValidateDiagFunc: ValidateRepoLayoutRefSchemaOverride,
		Description:      "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
	},
}

var ProxySchema = map[string]*schema.Schema{
	"proxy": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Proxy key from Artifactory Proxies settings. Can't be set if `disable_proxy = true`.",
	},
	"disable_proxy": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set to `true`, the proxy is disabled, and not returned in the API response body. If there is a default proxy set for the Artifactory instance, it will be ignored, too. Introduced since Artifactory 7.41.7.",
	},
}

var CompressionFormats = map[string]*schema.Schema{
	"index_compression_formats": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Set:      schema.HashString,
		Optional: true,
	},
}

var AlpinePrimaryKeyPairRef = map[string]*schema.Schema{
	"primary_keypair_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Description: "Used to sign index files in Alpine Linux repositories. " +
			"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
	},
}

var PrimaryKeyPairRef = map[string]*schema.Schema{
	"primary_keypair_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Primary keypair used to sign artifacts. Default value is empty.",
	},
}

var SecondaryKeyPairRef = map[string]*schema.Schema{
	"secondary_keypair_ref": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Secondary keypair used to sign artifacts.",
	},
}

type PrimaryKeyPairRefParam struct {
	PrimaryKeyPairRef string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
}

type SecondaryKeyPairRefParam struct {
	SecondaryKeyPairRef string `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef"`
}

type ContentSynchronisation struct {
	Enabled    bool                             `json:"enabled"`
	Statistics ContentSynchronisationStatistics `json:"statistics"`
	Properties ContentSynchronisationProperties `json:"properties"`
	Source     ContentSynchronisationSource     `json:"source"`
}

type ContentSynchronisationStatistics struct {
	Enabled bool `hcl:"statistics_enabled" json:"enabled"`
}

type ContentSynchronisationProperties struct {
	Enabled bool `hcl:"properties_enabled" json:"enabled"`
}

type ContentSynchronisationSource struct {
	OriginAbsenceDetection bool `hcl:"source_origin_absence_detection" json:"originAbsenceDetection"`
}

type ReadFunc func(d *schema.ResourceData, m interface{}) error

// Constructor Must return a pointer to a struct. When just returning a struct, resty gets confused and thinks it's a map
type Constructor func() (interface{}, error)

func Create(ctx context.Context, d *schema.ResourceData, m interface{}, unpack unpacker.UnpackFunc) diag.Diagnostics {
	repo, key, err := unpack(d)
	if err != nil {
		return diag.FromErr(err)
	}
	// repo must be a pointer
	res, err := m.(util.ProviderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetBody(repo).
		SetPathParam("key", key).
		Put(RepositoriesEndpoint)

	if err != nil {
		return diag.FromErr(err)
	}
	if res.IsError() {
		return diag.Errorf("%s", res.String())
	}

	d.SetId(key)

	return nil
}

func MkRepoCreate(unpack unpacker.UnpackFunc, read schema.ReadContextFunc) schema.CreateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		err := Create(ctx, d, m, unpack)
		if err != nil {
			return err
		}

		return read(ctx, d, m)
	}
}

func MkRepoRead(pack packer.PackFunc, construct Constructor) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, err := construct()
		if err != nil {
			return diag.FromErr(err)
		}

		// repo must be a pointer
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetResult(repo).
			SetPathParam("key", d.Id()).
			Get(RepositoriesEndpoint)

		if err != nil {
			return diag.FromErr(err)
		}
		if resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		if resp.IsError() {
			return diag.Errorf("%s", resp.String())
		}

		return diag.FromErr(pack(repo, d))
	}
}

func Update(ctx context.Context, d *schema.ResourceData, m interface{}, unpack unpacker.UnpackFunc) diag.Diagnostics {
	repo, key, err := unpack(d)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := m.(util.ProviderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetBody(repo).
		SetPathParam("key", d.Id()).
		Post(RepositoriesEndpoint)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp.IsError() {
		return diag.Errorf("%s", resp.String())
	}

	d.SetId(key)

	projectKeyChanged := d.HasChange("project_key")
	if projectKeyChanged {
		old, newProject := d.GetChange("project_key")
		oldProjectKey := old.(string)
		newProjectKey := newProject.(string)

		assignToProject := oldProjectKey == "" && len(newProjectKey) > 0
		unassignFromProject := len(oldProjectKey) > 0 && newProjectKey == ""

		var err error
		if assignToProject {
			err = assignRepoToProject(key, newProjectKey, m.(util.ProviderMetadata).Client)
		} else if unassignFromProject {
			err = unassignRepoFromProject(key, m.(util.ProviderMetadata).Client)
		}

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func MkRepoUpdate(unpack unpacker.UnpackFunc, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		err := Update(ctx, d, m, unpack)
		if err != nil {
			return err
		}

		return read(ctx, d, m)
	}
}

func assignRepoToProject(repoKey string, projectKey string, client *resty.Client) error {
	_, err := client.R().
		SetPathParams(map[string]string{
			"repoKey":    repoKey,
			"projectKey": projectKey,
		}).
		Put("access/api/v1/projects/_/attach/repositories/{repoKey}/{projectKey}")
	return err
}

func unassignRepoFromProject(repoKey string, client *resty.Client) error {
	_, err := client.R().
		SetPathParam("repoKey", repoKey).
		Delete("access/api/v1/projects/_/attach/repositories/{repoKey}")
	return err
}

type RepositoryFileList struct {
	URI   string            `json:"uri"`
	Files []json.RawMessage `json:"files"`
}

func GetArtifactCount(repoKey string, client *resty.Client) (int, error) {
	var fileList RepositoryFileList

	resp, err := client.R().
		SetPathParam("repo_key", repoKey).
		SetQueryParams(map[string]string{
			"list":        "",
			"deep":        "1",
			"listFolders": "0",
		}).
		SetResult(&fileList).
		Get("artifactory/api/storage/{repo_key}")

	if err != nil {
		return -1, err
	}

	if resp.IsError() {
		return -1, fmt.Errorf("%s", resp.String())
	}

	return len(fileList.Files), nil
}

func DeleteRepo(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(util.ProviderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetPathParam("key", d.Id()).
		Delete(RepositoriesEndpoint)

	if err != nil {
		return diag.FromErr(err)
	}

	if resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if resp.IsError() {
		return diag.Errorf("%s", resp.String())
	}

	return nil
}

func Retry400(response *resty.Response, _ error) bool {
	return response.StatusCode() == http.StatusBadRequest
}

var PackageTypesLikeGradle = []string{
	GradlePackageType,
	SBTPackageType,
	IvyPackageType,
}

var ProjectEnvironmentsSupported = []string{"DEV", "PROD"}

func RepoLayoutRefSchema(repositoryType string, packageType string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"repo_layout_ref": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: GetDefaultRepoLayoutRef(repositoryType, packageType),
			Description: fmt.Sprintf("Repository layout key for the %s repository", repositoryType),
		},
	}
}

// HandleResetWithNonExistentValue Special handling for field that requires non-existant value for RT
//
// Artifactory REST API will not accept empty string or null to reset value to not set
// Instead, using a non-existant value works as a workaround
// To ensure we don't accidentally set the value to a valid value, we use a UUID v4 string
func HandleResetWithNonExistentValue(d *utilsdk.ResourceData, key string) string {
	value := d.GetString(key, false)

	// When value has changed and is empty string, then it has been removed from
	// the Terraform configuration.
	if value == "" && d.HasChange(key) {
		return fmt.Sprintf("non-existant-value-%d", testutil.RandomInt())
	}

	return value
}

const CustomProjectEnvironmentSupportedVersion = "7.53.1"

func ProjectEnvironmentsDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if data, ok := diff.GetOk("project_environments"); ok {
		projectEnvironments := data.(*schema.Set).List()
		providerMetadata := meta.(util.ProviderMetadata)

		isSupported, err := util.CheckVersion(providerMetadata.ArtifactoryVersion, CustomProjectEnvironmentSupportedVersion)
		if err != nil {
			return fmt.Errorf("failed to check version %s", err)
		}

		if isSupported {
			if len(projectEnvironments) == 2 {
				return fmt.Errorf("for Artifactory %s or later, only one environment can be assigned to a repository", CustomProjectEnvironmentSupportedVersion)
			}
		} else { // Before 7.53.1
			projectEnvironments := data.(*schema.Set).List()
			for _, projectEnvironment := range projectEnvironments {
				if !slices.Contains(ProjectEnvironmentsSupported, projectEnvironment.(string)) {
					return fmt.Errorf("project_environment %s not allowed", projectEnvironment)
				}
			}
		}
	}

	return nil
}

func VerifyDisableProxy(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	disableProxy := diff.Get("disable_proxy").(bool)
	proxy := diff.Get("proxy").(string)

	if disableProxy && len(proxy) > 0 {
		return fmt.Errorf("if `disable_proxy` is set to `true`, `proxy` can't be set")
	}

	return nil
}

func MkResourceSchema(skeemas map[int16]map[string]*schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor Constructor) *schema.Resource {
	var reader = MkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: MkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: MkRepoUpdate(unpack, reader),
		DeleteContext: DeleteRepo,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        skeemas[1],
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				// this only works because the schema hasn't changed, except the removal of default value
				// from `project_key` attribute. Future common schema changes that involve attributes should
				// figure out a way to create a previous and new version.
				Type:    Resource(skeemas[0]).CoreConfigSchema().ImpliedType(),
				Upgrade: ResourceUpgradeProjectKey,
				Version: 0,
			},
		},

		CustomizeDiff: ProjectEnvironmentsDiff,
	}
}

func Resource(skeema map[string]*schema.Schema) *schema.Resource {
	return &schema.Resource{
		Schema: skeema,
	}
}

func ResourceUpgradeProjectKey(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState["project_key"] == "default" {
		rawState["project_key"] = ""
	}

	return rawState, nil
}

const RepositoriesEndpoint = "artifactory/api/repositories/{key}"

func CheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return request.SetPathParam("key", id).Head(RepositoriesEndpoint)
}

func ValidateRepoLayoutRefSchemaOverride(_ interface{}, _ cty.Path) diag.Diagnostics {
	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Always override repo_layout_ref attribute in the schema",
			Detail:   "Always override repo_layout_ref attribute in the schema on top of base schema",
		},
	}
}

type SupportedRepoClasses struct {
	RepoLayoutRef      string
	SupportedRepoTypes map[string]bool
}

// GetDefaultRepoLayoutRef return the default repo layout by Repository Type & Package Type
func GetDefaultRepoLayoutRef(repositoryType, packageType string) func() (interface{}, error) {
	return func() (interface{}, error) {
		if v, ok := defaultRepoLayoutMap[packageType].SupportedRepoTypes[repositoryType]; ok && v {
			return defaultRepoLayoutMap[packageType].RepoLayoutRef, nil
		}
		return nil, fmt.Errorf("default repo layout not found for repository type %s & package type %s", repositoryType, packageType)
	}
}
