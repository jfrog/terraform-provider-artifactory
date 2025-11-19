package virtual

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

const (
	Rclass               = "virtual"
	CurrentSchemaVersion = 1
)

// Framework Support

func NewVirtualRepositoryResource(packageType, packageName string, resourceModelType, apiModelType reflect.Type) virtualResource {
	return virtualResource{
		BaseResource: repository.NewRepositoryResource(packageType, packageName, Rclass, resourceModelType, apiModelType),
	}
}

type virtualResource struct {
	repository.BaseResource
}

type VirtualResourceModel struct {
	repository.BaseResourceModel
	Repositories                                  types.List   `tfsdk:"repositories"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts types.Bool   `tfsdk:"artifactory_requests_can_retrieve_remote_artifacts"`
	DefaultDeploymentRepo                         types.String `tfsdk:"default_deployment_repo"`
	RepoLayoutRef                                 types.String `tfsdk:"repo_layout_ref"`
}

func (r *VirtualResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r VirtualResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *VirtualResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r VirtualResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *VirtualResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *VirtualResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r VirtualResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r VirtualResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// Get base API model
	baseModel, d := r.BaseResourceModel.ToAPIModel(ctx, "virtual", packageType)
	if d != nil {
		diags.Append(d...)
	}

	var repositories []string
	d = r.Repositories.ElementsAs(ctx, &repositories, false)
	if d != nil {
		diags.Append(d...)
	}

	// Handle repository layout reference
	repoLayoutRef := r.RepoLayoutRef.ValueString()
	if r.RepoLayoutRef.IsNull() || repoLayoutRef == "" {
		defaultRepoLayout, err := repository.GetDefaultRepoLayoutRef("virtual", packageType)
		if err != nil {
			diags.AddError(
				"Failed to get default repo layout ref",
				err.Error(),
			)
		} else {
			repoLayoutRef = defaultRepoLayout
		}
	}

	return VirtualAPIModel{
		BaseAPIModel: baseModel.(repository.BaseAPIModel),
		Repositories: repositories,
		ArtifactoryRequestsCanRetrieveRemoteArtifacts: r.ArtifactoryRequestsCanRetrieveRemoteArtifacts.ValueBool(),
		DefaultDeploymentRepo:                         r.DefaultDeploymentRepo.ValueString(),
		RepoLayoutRef:                                 repoLayoutRef,
	}, diags
}

func (r *VirtualResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(VirtualAPIModel)

	// Set base model fields
	r.BaseResourceModel.FromAPIModel(ctx, model.BaseAPIModel)

	// Set virtual-specific fields
	r.ArtifactoryRequestsCanRetrieveRemoteArtifacts = types.BoolValue(model.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	r.DefaultDeploymentRepo = types.StringValue(model.DefaultDeploymentRepo)
	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)

	repositories, ds := types.ListValueFrom(ctx, types.StringType, model.Repositories)
	if ds.HasError() {
		diags.Append(ds...)
		return diags
	}
	r.Repositories = repositories

	return diags
}

type VirtualAPIModel struct {
	repository.BaseAPIModel
	Repositories                                  []string `json:"repositories,omitempty"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts bool     `json:"artifactoryRequestsCanRetrieveRemoteArtifacts"`
	DefaultDeploymentRepo                         string   `json:"defaultDeploymentRepo,omitempty"`
	RepoLayoutRef                                 string   `json:"repoLayoutRef,omitempty"`
}

var VirtualAttributes = lo.Assign(
	repository.BaseAttributes,
	map[string]schema.Attribute{
		"repositories": schema.ListAttribute{
			ElementType:         types.StringType,
			Optional:            true,
			Computed:            true,
			Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			MarkdownDescription: "The effective list of actual repositories included in this virtual repository.",
		},
		"artifactory_requests_can_retrieve_remote_artifacts": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Whether the virtual repository should search through remote repositories when trying to resolve an artifact requested by another Artifactory instance.",
		},
		"default_deployment_repo": schema.StringAttribute{
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			MarkdownDescription: "Default repository to deploy artifacts.",
		},
	},
)

type RepositoryBaseParams struct {
	Key                                           string   `hcl:"key" json:"key,omitempty"`
	ProjectKey                                    string   `json:"projectKey"`
	ProjectEnvironments                           []string `json:"environments"`
	Rclass                                        string   `json:"rclass"`
	PackageType                                   string   `hcl:"package_type" json:"packageType,omitempty"`
	Description                                   string   `json:"description"`
	Notes                                         string   `json:"notes"`
	IncludesPattern                               string   `json:"includesPattern"`
	ExcludesPattern                               string   `json:"excludesPattern"`
	RepoLayoutRef                                 string   `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	Repositories                                  []string `hcl:"repositories" json:"repositories,omitempty"`
	ArtifactoryRequestsCanRetrieveRemoteArtifacts bool     `hcl:"artifactory_requests_can_retrieve_remote_artifacts" json:"artifactoryRequestsCanRetrieveRemoteArtifacts"`
	DefaultDeploymentRepo                         string   `hcl:"default_deployment_repo" json:"defaultDeploymentRepo,omitempty"`
}

type RepositoryBaseParamsWithRetrievalCachePeriodSecs struct {
	RepositoryBaseParams
	VirtualRetrievalCachePeriodSecs int `hcl:"retrieval_cache_period_seconds" json:"virtualRetrievalCachePeriodSecs"`
}

func (bp RepositoryBaseParams) Id() string {
	return bp.Key
}

var PackageTypesLikeGeneric = []string{
	repository.CocoapodsPackageType,
	repository.ComposerPackageType,
	repository.GemsPackageType,
	repository.GenericPackageType,
	repository.GitLFSPackageType,
	repository.P2PackageType,
	repository.PubPackageType,
	repository.PuppetPackageType,
	repository.PyPiPackageType,
	repository.SwiftPackageType,
	repository.TerraformPackageType,
}

var PackageTypesLikeGenericWithRetrievalCachePeriodSecs = []string{
	repository.AnsiblePackageType,
	repository.ChefPackageType,
	repository.CondaPackageType,
	repository.CranPackageType,
}

var baseSchema = map[string]*sdkv2_schema.Schema{
	"repositories": {
		Type:        sdkv2_schema.TypeList,
		Elem:        &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
		Optional:    true,
		Description: "The effective list of actual repositories included in this virtual repository.",
	},
	"artifactory_requests_can_retrieve_remote_artifacts": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Whether the virtual repository should search through remote repositories when trying to resolve an artifact requested by another Artifactory instance.",
	},
	"default_deployment_repo": {
		Type:        sdkv2_schema.TypeString,
		Optional:    true,
		Description: "Default repository to deploy artifacts.",
	},
}

var BaseSchemaV1 = lo.Assign(
	repository.BaseSchemaV1,
	baseSchema,
)

var GetSchemas = func(s map[string]*sdkv2_schema.Schema) map[int16]map[string]*sdkv2_schema.Schema {
	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(
			BaseSchemaV1,
			s,
		),
		1: lo.Assign(
			BaseSchemaV1,
			s,
		),
	}
}

func UnpackBaseVirtRepo(s *sdkv2_schema.ResourceData, packageType string) RepositoryBaseParams {
	d := &utilsdk.ResourceData{ResourceData: s}

	return RepositoryBaseParams{
		Key:                 d.GetString("key", false),
		Rclass:              Rclass,
		ProjectKey:          d.GetString("project_key", false),
		ProjectEnvironments: d.GetSet("project_environments"),
		PackageType:         packageType, // must be set independently
		IncludesPattern:     d.GetString("includes_pattern", false),
		ExcludesPattern:     d.GetString("excludes_pattern", false),
		RepoLayoutRef:       d.GetString("repo_layout_ref", false),
		ArtifactoryRequestsCanRetrieveRemoteArtifacts: d.GetBool("artifactory_requests_can_retrieve_remote_artifacts", false),
		Repositories:          d.GetList("repositories"),
		Description:           d.GetString("description", false),
		Notes:                 d.GetString("notes", false),
		DefaultDeploymentRepo: repository.HandleResetWithNonExistentValue(d, "default_deployment_repo"),
	}
}

func UnpackBaseVirtRepoWithRetrievalCachePeriodSecs(s *sdkv2_schema.ResourceData, packageType string) RepositoryBaseParamsWithRetrievalCachePeriodSecs {
	d := &utilsdk.ResourceData{ResourceData: s}

	return RepositoryBaseParamsWithRetrievalCachePeriodSecs{
		RepositoryBaseParams:            UnpackBaseVirtRepo(s, packageType),
		VirtualRetrievalCachePeriodSecs: d.GetInt("retrieval_cache_period_seconds", false),
	}
}

var externalDependenciesSchema = map[string]*sdkv2_schema.Schema{
	"external_dependencies_enabled": {
		Type:        sdkv2_schema.TypeBool,
		Default:     false,
		Optional:    true,
		Description: "When set, external dependencies are rewritten. Default value is false.",
	},
	"external_dependencies_remote_repo": {
		Type:             sdkv2_schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		RequiredWith:     []string{"external_dependencies_enabled"},
		Description:      "The remote repository aggregated by this virtual repository in which the external dependency will be cached.",
	},
	"external_dependencies_patterns": {
		Type:     sdkv2_schema.TypeList,
		Optional: true,
		Elem: &sdkv2_schema.Schema{
			Type: sdkv2_schema.TypeString,
		},
		RequiredWith: []string{"external_dependencies_enabled"},
		Description: "An Allow List of Ant-style path expressions that specify where external dependencies may be downloaded from. " +
			"By default, this is set to ** which means that dependencies may be downloaded from any external source.",
	},
}

type ExternalDependenciesVirtualRepositoryParams struct {
	RepositoryBaseParams
	ExternalDependenciesEnabled    bool     `json:"externalDependenciesEnabled"`
	ExternalDependenciesRemoteRepo string   `json:"externalDependenciesRemoteRepo"`
	ExternalDependenciesPatterns   []string `json:"externalDependenciesPatterns"`
}

var unpackExternalDependenciesVirtualRepository = func(s *sdkv2_schema.ResourceData, packageType string) ExternalDependenciesVirtualRepositoryParams {
	d := &utilsdk.ResourceData{ResourceData: s}

	return ExternalDependenciesVirtualRepositoryParams{
		RepositoryBaseParams:           UnpackBaseVirtRepo(s, packageType),
		ExternalDependenciesEnabled:    d.GetBool("external_dependencies_enabled", false),
		ExternalDependenciesRemoteRepo: d.GetString("external_dependencies_remote_repo", false),
		ExternalDependenciesPatterns:   d.GetList("external_dependencies_patterns"),
	}
}

var RetrievalCachePeriodSecondsSchema = map[string]*sdkv2_schema.Schema{
	"retrieval_cache_period_seconds": {
		Type:     sdkv2_schema.TypeInt,
		Optional: true,
		Default:  7200,
		Description: "This value refers to the number of seconds to cache metadata files before checking for newer " +
			"versions on aggregated repositories. A value of 0 indicates no caching.",
		ValidateFunc: validation.IntAtLeast(0),
	},
}
