package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdkv2_diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/unpacker"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	sdkv2_validator "github.com/jfrog/terraform-provider-shared/validator"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"

	"golang.org/x/exp/slices"
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
	MachineLearningType          = "machinelearning"
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

type BaseResource struct {
	util.JFrogResource
	Description       string
	PackageType       string
	Rclass            string
	ResourceModelType reflect.Type
	APIModelType      reflect.Type
}

func (r BaseResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.ProviderData == nil {
		return
	}

	var projectEnviroments types.Set

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("project_environments"), &projectEnviroments)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If project_environments is not configured, return without warning.
	if projectEnviroments.IsNull() || projectEnviroments.IsUnknown() {
		return
	}

	var envs []string
	resp.Diagnostics.Append(projectEnviroments.ElementsAs(ctx, &envs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isSupported, err := util.CheckVersion(r.ProviderData.ArtifactoryVersion, CustomProjectEnvironmentSupportedVersion)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to check Artifactory version",
			err.Error(),
		)
		return
	}

	if isSupported {
		if len(envs) == 2 {
			resp.Diagnostics.AddError(
				"Too many project environment",
				fmt.Sprintf("for Artifactory %s or later, only one environment can be assigned to a repository", CustomProjectEnvironmentSupportedVersion),
			)
			return
		}
	} else { // Before 7.53.1
		for _, env := range envs {
			if !slices.Contains(ProjectEnvironmentsSupported, env) {
				resp.Diagnostics.AddError(
					"Invalid project_environment not allowed",
					env,
				)
				return
			}
		}
	}
}

func (r *BaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.ResourceModelType == nil || r.APIModelType == nil {
		resp.Diagnostics.AddError(
			"ResourceModelType or APIModelType is nil",
			"",
		)
		return
	}

	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	plan := reflect.New(r.ResourceModelType).Interface().(ResourceModelIface)

	plan.GetCreateResourcePlanData(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, d := plan.ToAPIModel(ctx, r.PackageType)
	if d != nil {
		resp.Diagnostics.Append(d...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("key", plan.KeyString()).
		SetBody(repo).
		SetError(&jfrogErrors).
		Put(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, jfrogErrors.String())
		return
	}

	plan.SetID(plan.KeyString())

	if plan.ProjectEnvironmentsValue().IsUnknown() {
		plan.SetProjectEnvironments(types.SetNull(types.StringType))
	}

	plan.SetCreateResourceStateData(ctx, resp)
}

func (r *BaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	state := reflect.New(r.ResourceModelType).Interface().(ResourceModelIface)

	state.GetReadResourceStateData(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	repo := reflect.New(r.APIModelType).Interface()
	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("key", state.KeyString()).
		SetResult(repo).
		SetError(&jfrogErrors).
		Get(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	// Treat HTTP 404 Not Found status as a signal to recreate resource
	// and return early
	if response.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, jfrogErrors.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.FromAPIModel(ctx, repo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.SetReadResourceStateData(ctx, resp)
}

func (r *BaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	plan := reflect.New(r.ResourceModelType).Interface().(ResourceModelIface)
	plan.GetUpdateResourcePlanData(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	state := reflect.New(r.ResourceModelType).Interface().(ResourceModelIface)
	state.GetUpdateResourceStateData(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	repo, d := plan.ToAPIModel(ctx, r.PackageType)
	if d != nil {
		resp.Diagnostics.Append(d...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	key := plan.KeyString()
	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("key", key).
		SetBody(repo).
		SetError(&jfrogErrors).
		Post(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, jfrogErrors.String())
		return
	}

	if plan.ProjectEnvironmentsValue().IsUnknown() {
		plan.SetProjectEnvironments(types.SetNull(types.StringType))
	}

	plan.SetID(plan.KeyString())

	planProjectKey := plan.ProjectKeyValue()
	stateProjectKey := state.ProjectKeyValue()

	if !planProjectKey.Equal(stateProjectKey) {
		oldProjectKey := stateProjectKey.ValueString()
		newProjectKey := planProjectKey.ValueString()

		assignToProject := oldProjectKey == "" && len(newProjectKey) > 0
		unassignFromProject := len(oldProjectKey) > 0 && newProjectKey == ""

		var err error
		if assignToProject {
			err = AssignRepoToProject(key, newProjectKey, r.ProviderData.Client)
		} else if unassignFromProject {
			err = UnassignRepoFromProject(key, r.ProviderData.Client)
		}

		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to assign/unassign repository to project",
				err.Error(),
			)
			return
		}
	}

	plan.SetUpdateResourceStateData(ctx, resp)
}

func (r *BaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var key types.String

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("key"), &key)...)

	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("key", key.ValueString()).
		SetError(&jfrogErrors).
		Delete(r.DocumentEndpoint)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// Return error if the HTTP status code is not 200 OK
	if response.StatusCode() != http.StatusOK {
		utilfw.UnableToDeleteResourceError(resp, jfrogErrors.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *BaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

type ResourceModelIface interface {
	SetID(string)
	KeyString() string
	ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics)
	FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics
	GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)
	SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse)
	GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse)
	SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse)
	GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse)
	GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse)
	SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse)
	ProjectEnvironmentsValue() basetypes.SetValue
	SetProjectEnvironments(basetypes.SetValue)
	ProjectKeyValue() basetypes.StringValue
}

type BaseResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Key                 types.String `tfsdk:"key"`
	ProjectKey          types.String `tfsdk:"project_key"`
	ProjectEnvironments types.Set    `tfsdk:"project_environments"`
	Description         types.String `tfsdk:"description"`
	Notes               types.String `tfsdk:"notes"`
	IncludesPattern     types.String `tfsdk:"includes_pattern"`
	ExcludesPattern     types.String `tfsdk:"excludes_pattern"`
}

func (r *BaseResourceModel) SetID(id string) {
	r.ID = types.StringValue(id)
}

func (r BaseResourceModel) KeyString() string {
	return r.Key.ValueString()
}

func (r BaseResourceModel) ProjectEnvironmentsValue() basetypes.SetValue {
	return r.ProjectEnvironments
}

func (r *BaseResourceModel) SetProjectEnvironments(v basetypes.SetValue) {
	r.ProjectEnvironments = v
}

func (r BaseResourceModel) ProjectKeyValue() basetypes.StringValue {
	return r.ProjectKey
}

func (r BaseResourceModel) ToAPIModel(ctx context.Context, rclass, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var projectEnviroments []string
	if !r.ProjectEnvironments.IsUnknown() {
		d := r.ProjectEnvironments.ElementsAs(ctx, &projectEnviroments, false)
		if d != nil {
			diags.Append(d...)
		}
	}

	repoLayoutRef, err := GetDefaultRepoLayoutRef(rclass, packageType)
	if err != nil {
		diags.AddError(
			"Failed to get default repo layout ref",
			err.Error(),
		)
	}

	return BaseAPIModel{
		Key:                 r.Key.ValueString(),
		ProjectKey:          r.ProjectKey.ValueString(),
		ProjectEnvironments: projectEnviroments,
		Rclass:              rclass,
		PackageType:         packageType,
		Description:         r.Description.ValueString(),
		Notes:               r.Notes.ValueString(),
		IncludesPattern:     r.IncludesPattern.ValueString(),
		ExcludesPattern:     r.ExcludesPattern.ValueString(),
		RepoLayoutRef:       repoLayoutRef,
	}, diags
}

func (r *BaseResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(BaseAPIModel)

	r.ID = types.StringValue(model.Key)
	r.Key = types.StringValue(model.Key)
	r.ProjectKey = types.StringValue(model.ProjectKey)
	r.Description = types.StringValue(model.Description)
	r.Notes = types.StringValue(model.Notes)
	r.IncludesPattern = types.StringValue(model.IncludesPattern)
	r.ExcludesPattern = types.StringValue(model.ExcludesPattern)

	projectEnviroments := types.SetNull(types.StringType)
	if len(model.ProjectEnvironments) > 0 {
		envs, ds := types.SetValueFrom(ctx, types.StringType, model.ProjectEnvironments)
		if ds.HasError() {
			diags.Append(ds...)
			return diags
		}
		projectEnviroments = envs
	}

	r.ProjectEnvironments = projectEnviroments

	return diags
}

type BaseAPIModel struct {
	Key                 string   `json:"key"`
	ProjectKey          string   `json:"projectKey"`
	ProjectEnvironments []string `json:"environments,omitempty"`
	Rclass              string   `json:"rclass"`
	PackageType         string   `json:"packageType"`
	Description         string   `json:"description"`
	Notes               string   `json:"notes"`
	IncludesPattern     string   `json:"includesPattern"`
	ExcludesPattern     string   `json:"excludesPattern"`
	RepoLayoutRef       string   `json:"repoLayoutRef"`
}

var BaseAttributes = map[string]schema.Attribute{
	"id": schema.StringAttribute{
		Computed: true,
	},
	"key": schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			validatorfw_string.RepoKey(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		MarkdownDescription: "A mandatory identifier for the repository that must be unique. Must be 1 - 64 alphanumeric and hyphen characters. It cannot contain spaces or special characters.",
	},
	"project_key": schema.StringAttribute{
		Optional: true,
		Computed: true,
		Default:  stringdefault.StaticString(""),
		Validators: []validator.String{
			validatorfw_string.ProjectKey(),
		},
		MarkdownDescription: "Project key for assigning this repository to. Must be 2 - 32 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": schema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
		Validators: []validator.Set{
			setvalidator.SizeBetween(0, 2),
		},
		MarkdownDescription: "Project environment for assigning this repository to. Allow values: \"DEV\", \"PROD\", or one of custom environment. " +
			"Before Artifactory 7.53.1, up to 2 values (\"DEV\" and \"PROD\") are allowed. From 7.53.1 onward, only one value is allowed. " +
			"The attribute should only be used if the repository is already assigned to the existing project. If not, " +
			"the attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create " +
			"state drift during the update.",
	},
	"description": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
		MarkdownDescription: "Public description.",
	},
	"notes": schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		Default:             stringdefault.StaticString(""),
		MarkdownDescription: "Internal description.",
	},
	"includes_pattern": schema.StringAttribute{
		Optional: true,
		Computed: true,
		Default:  stringdefault.StaticString("**/*"),
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
		MarkdownDescription: "List of comma-separated artifact patterns to include when evaluating artifact requests in the form of `x/y/**/z/*`. " +
			"When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (`**/*`).",
	},
	"excludes_pattern": schema.StringAttribute{
		Optional: true,
		Computed: true,
		Default:  stringdefault.StaticString(""),
		MarkdownDescription: "List of artifact patterns to exclude when evaluating artifact requests, in the form of `x/y/**/z/*`." +
			"By default no artifacts are excluded.",
	},
}

func RepoLayoutRefAttribute(repositoryType string, packageType string) map[string]schema.Attribute {
	var defaultRepoLayout string
	if v, ok := defaultRepoLayoutMap[packageType].SupportedRepoTypes[repositoryType]; ok && v {
		defaultRepoLayout = defaultRepoLayoutMap[packageType].RepoLayoutRef
	}

	return map[string]schema.Attribute{
		"repo_layout_ref": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(defaultRepoLayout),
			Description: "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
		},
	}
}

var PrimaryKeyPairRefAttribute = map[string]schema.Attribute{
	"primary_keypair_ref": schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
		MarkdownDescription: "Primary keypair used to sign artifacts. Default value is empty.",
	},
}

var SecondaryKeyPairRefAttribute = map[string]schema.Attribute{
	"secondary_keypair_ref": schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
		MarkdownDescription: "Secondary keypair used to sign artifacts.",
	},
}

var CompressionFormatsAttribute = map[string]schema.Attribute{
	"index_compression_formats": schema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
	},
}

var BaseSchemaV1 = map[string]*sdkv2_schema.Schema{
	"key": {
		Type:             sdkv2_schema.TypeString,
		Required:         true,
		ForceNew:         true,
		ValidateDiagFunc: sdkv2_validator.RepoKey,
		Description:      "A mandatory identifier for the repository that must be unique. Must be 1 - 64 alphanumeric and hyphen characters. It cannot contain spaces or special characters.",
	},
	"project_key": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: sdkv2_validator.ProjectKey,
		Description:      "Project key for assigning this repository to. Must be 2 - 32 lowercase alphanumeric and hyphen characters. When assigning repository to a project, repository key must be prefixed with project key, separated by a dash.",
	},
	"project_environments": {
		Type:     sdkv2_schema.TypeSet,
		Elem:     &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
		MinItems: 0,
		MaxItems: 2,
		Set:      sdkv2_schema.HashString,
		Optional: true,
		Computed: true,
		Description: "Project environment for assigning this repository to. Allow values: \"DEV\", \"PROD\", or one of custom environment. " +
			"Before Artifactory 7.53.1, up to 2 values (\"DEV\" and \"PROD\") are allowed. From 7.53.1 onward, only one value is allowed. " +
			"The attribute should only be used if the repository is already assigned to the existing project. If not, " +
			"the attribute will be ignored by Artifactory, but will remain in the Terraform state, which will create " +
			"state drift during the update.",
	},
	"package_type": {
		Type:     sdkv2_schema.TypeString,
		Required: false,
		Computed: true,
		ForceNew: true,
	},
	"description": {
		Type:        sdkv2_schema.TypeString,
		Optional:    true,
		Description: "Public description.",
	},
	"notes": {
		Type:        sdkv2_schema.TypeString,
		Optional:    true,
		Description: "Internal description.",
	},
	"includes_pattern": {
		Type:     sdkv2_schema.TypeString,
		Optional: true,
		Default:  "**/*",
		Description: "List of comma-separated artifact patterns to include when evaluating artifact requests in the form of `x/y/**/z/*`. " +
			"When used, only artifacts matching one of the include patterns are served. By default, all artifacts are included (`**/*`).",
	},
	"excludes_pattern": {
		Type:     sdkv2_schema.TypeString,
		Optional: true,
		Description: "List of artifact patterns to exclude when evaluating artifact requests, in the form of `x/y/**/z/*`." +
			"By default no artifacts are excluded.",
	},
	"repo_layout_ref": {
		Type:     sdkv2_schema.TypeString,
		Optional: true,
		// The default value in the UI is simple-default, in API maven-2-default. Provider will always override it ro math the UI.
		ValidateDiagFunc: ValidateRepoLayoutRefSchemaOverride,
		Description:      "Sets the layout that the repository should use for storing and identifying modules. A recommended layout that corresponds to the package type defined is suggested, and index packages uploaded and calculate metadata accordingly.",
	},
}

var ProxySchemaSDKv2 = map[string]*sdkv2_schema.Schema{
	"proxy": {
		Type:        sdkv2_schema.TypeString,
		Optional:    true,
		Description: "Proxy key from Artifactory Proxies settings. Can't be set if `disable_proxy = true`.",
	},
	"disable_proxy": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set to `true`, the proxy is disabled, and not returned in the API response body. If there is a default proxy set for the Artifactory instance, it will be ignored, too. Introduced since Artifactory 7.41.7.",
	},
}

var CompressionFormatsSDKv2 = map[string]*sdkv2_schema.Schema{
	"index_compression_formats": {
		Type: sdkv2_schema.TypeSet,
		Elem: &sdkv2_schema.Schema{
			Type: sdkv2_schema.TypeString,
		},
		Set:      sdkv2_schema.HashString,
		Optional: true,
	},
}

var AlpinePrimaryKeyPairRefSDKv2 = map[string]*sdkv2_schema.Schema{
	"primary_keypair_ref": {
		Type:     sdkv2_schema.TypeString,
		Optional: true,
		Description: "Used to sign index files in Alpine Linux repositories. " +
			"See: https://www.jfrog.com/confluence/display/JFROG/Alpine+Linux+Repositories#AlpineLinuxRepositories-SigningAlpineLinuxIndex",
	},
}

var PrimaryKeyPairRefSDKv2 = map[string]*sdkv2_schema.Schema{
	"primary_keypair_ref": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Primary keypair used to sign artifacts. Default value is empty.",
	},
}

var SecondaryKeyPairRefSDKv2 = map[string]*sdkv2_schema.Schema{
	"secondary_keypair_ref": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Secondary keypair used to sign artifacts.",
	},
}

type PrimaryKeyPairRefParam struct {
	PrimaryKeyPairRefSDKv2 string `hcl:"primary_keypair_ref" json:"primaryKeyPairRef"`
}

type SecondaryKeyPairRefParam struct {
	SecondaryKeyPairRefSDKv2 string `hcl:"secondary_keypair_ref" json:"secondaryKeyPairRef"`
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

type ReadFunc func(d *sdkv2_schema.ResourceData, m interface{}) error

// Constructor Must return a pointer to a struct. When just returning a struct, resty gets confused and thinks it's a map
type Constructor func() (interface{}, error)

func Create(ctx context.Context, d *sdkv2_schema.ResourceData, m interface{}, unpack unpacker.UnpackFunc) sdkv2_diag.Diagnostics {
	repo, key, err := unpack(d)
	if err != nil {
		return sdkv2_diag.FromErr(err)
	}
	// repo must be a pointer
	res, err := m.(util.ProviderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetBody(repo).
		SetPathParam("key", key).
		Put(RepositoriesEndpoint)

	if err != nil {
		return sdkv2_diag.FromErr(err)
	}
	if res.IsError() {
		return sdkv2_diag.Errorf("%s", res.String())
	}

	d.SetId(key)

	return nil
}

func MkRepoCreate(unpack unpacker.UnpackFunc, read sdkv2_schema.ReadContextFunc) sdkv2_schema.CreateContextFunc {
	return func(ctx context.Context, d *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		err := Create(ctx, d, m, unpack)
		if err != nil {
			return err
		}

		return read(ctx, d, m)
	}
}

func MkRepoRead(pack packer.PackFunc, construct Constructor) sdkv2_schema.ReadContextFunc {
	return func(ctx context.Context, d *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		repo, err := construct()
		if err != nil {
			return sdkv2_diag.FromErr(err)
		}

		// repo must be a pointer
		resp, err := m.(util.ProviderMetadata).Client.R().
			SetResult(repo).
			SetPathParam("key", d.Id()).
			Get(RepositoriesEndpoint)

		if err != nil {
			return sdkv2_diag.FromErr(err)
		}
		if resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		if resp.IsError() {
			return sdkv2_diag.Errorf("%s", resp.String())
		}

		return sdkv2_diag.FromErr(pack(repo, d))
	}
}

func Update(ctx context.Context, d *sdkv2_schema.ResourceData, m interface{}, unpack unpacker.UnpackFunc) sdkv2_diag.Diagnostics {
	repo, key, err := unpack(d)
	if err != nil {
		return sdkv2_diag.FromErr(err)
	}

	resp, err := m.(util.ProviderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetBody(repo).
		SetPathParam("key", d.Id()).
		Post(RepositoriesEndpoint)
	if err != nil {
		return sdkv2_diag.FromErr(err)
	}

	if resp.IsError() {
		return sdkv2_diag.Errorf("%s", resp.String())
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
			err = AssignRepoToProject(key, newProjectKey, m.(util.ProviderMetadata).Client)
		} else if unassignFromProject {
			err = UnassignRepoFromProject(key, m.(util.ProviderMetadata).Client)
		}

		if err != nil {
			return sdkv2_diag.FromErr(err)
		}
	}

	return nil
}

func MkRepoUpdate(unpack unpacker.UnpackFunc, read sdkv2_schema.ReadContextFunc) sdkv2_schema.UpdateContextFunc {
	return func(ctx context.Context, d *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
		err := Update(ctx, d, m, unpack)
		if err != nil {
			return err
		}

		return read(ctx, d, m)
	}
}

func AssignRepoToProject(repoKey string, projectKey string, client *resty.Client) error {
	_, err := client.R().
		SetPathParams(map[string]string{
			"repoKey":    repoKey,
			"projectKey": projectKey,
		}).
		Put("access/api/v1/projects/_/attach/repositories/{repoKey}/{projectKey}")
	return err
}

func UnassignRepoFromProject(repoKey string, client *resty.Client) error {
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

func DeleteRepo(ctx context.Context, d *sdkv2_schema.ResourceData, m interface{}) sdkv2_diag.Diagnostics {
	resp, err := m.(util.ProviderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		SetPathParam("key", d.Id()).
		Delete(RepositoriesEndpoint)

	if err != nil {
		return sdkv2_diag.FromErr(err)
	}

	if resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if resp.IsError() {
		return sdkv2_diag.Errorf("%s", resp.String())
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

func RepoLayoutRefSDKv2Schema(repositoryType string, packageType string) map[string]*sdkv2_schema.Schema {
	return map[string]*sdkv2_schema.Schema{
		"repo_layout_ref": {
			Type:     sdkv2_schema.TypeString,
			Optional: true,
			DefaultFunc: func() (interface{}, error) {
				var ref interface{}
				ref, err := GetDefaultRepoLayoutRef(repositoryType, packageType)
				return ref, err
			},
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

func ProjectEnvironmentsDiff(ctx context.Context, diff *sdkv2_schema.ResourceDiff, meta interface{}) error {
	if data, ok := diff.GetOk("project_environments"); ok {
		projectEnvironments := data.(*sdkv2_schema.Set).List()
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
			projectEnvironments := data.(*sdkv2_schema.Set).List()
			for _, projectEnvironment := range projectEnvironments {
				if !slices.Contains(ProjectEnvironmentsSupported, projectEnvironment.(string)) {
					return fmt.Errorf("project_environment %s not allowed", projectEnvironment)
				}
			}
		}
	}

	return nil
}

func VerifyDisableProxy(_ context.Context, diff *sdkv2_schema.ResourceDiff, _ interface{}) error {
	disableProxy := diff.Get("disable_proxy").(bool)
	proxy := diff.Get("proxy").(string)

	if disableProxy && len(proxy) > 0 {
		return fmt.Errorf("if `disable_proxy` is set to `true`, `proxy` can't be set")
	}

	return nil
}

func MkResourceSchema(skeemas map[int16]map[string]*sdkv2_schema.Schema, packer packer.PackFunc, unpack unpacker.UnpackFunc, constructor Constructor) *sdkv2_schema.Resource {
	var reader = MkRepoRead(packer, constructor)
	return &sdkv2_schema.Resource{
		CreateContext: MkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: MkRepoUpdate(unpack, reader),
		DeleteContext: DeleteRepo,

		Importer: &sdkv2_schema.ResourceImporter{
			StateContext: sdkv2_schema.ImportStatePassthroughContext,
		},

		Schema:        skeemas[1],
		SchemaVersion: 1,
		StateUpgraders: []sdkv2_schema.StateUpgrader{
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

func Resource(skeema map[string]*sdkv2_schema.Schema) *sdkv2_schema.Resource {
	return &sdkv2_schema.Resource{
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

func ValidateRepoLayoutRefSchemaOverride(_ interface{}, _ cty.Path) sdkv2_diag.Diagnostics {
	return sdkv2_diag.Diagnostics{
		sdkv2_diag.Diagnostic{
			Severity: sdkv2_diag.Error,
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
func GetDefaultRepoLayoutRef(repositoryType, packageType string) (string, error) {
	if v, ok := defaultRepoLayoutMap[packageType].SupportedRepoTypes[repositoryType]; ok && v {
		return defaultRepoLayoutMap[packageType].RepoLayoutRef, nil
	}
	return "", fmt.Errorf("default repo layout not found for repository type %s & package type %s", repositoryType, packageType)
}
