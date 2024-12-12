package local

import (
	"context"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

const (
	Rclass               = "local"
	CurrentSchemaVersion = 1
)

var PackageTypesLikeGeneric = []string{
	repository.BowerPackageType,
	repository.ChefPackageType,
	repository.CocoapodsPackageType,
	repository.ComposerPackageType,
	repository.CondaPackageType,
	repository.CranPackageType,
	repository.GemsPackageType,
	repository.GenericPackageType,
	repository.GitLFSPackageType,
	repository.GoPackageType,
	repository.HelmPackageType,
	repository.HuggingFacePackageType,
	repository.NPMPackageType,
	repository.OpkgPackageType,
	repository.PubPackageType,
	repository.PuppetPackageType,
	repository.PyPiPackageType,
	repository.SwiftPackageType,
	repository.TerraformBackendPackageType,
	repository.VagrantPackageType,
}

type localResource struct {
	repository.BaseResource
}

func (r *localResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repo LocalAPIModel
	resp.Diagnostics.Append(plan.ToAPIModel(ctx, r.PackageType, &repo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("key", plan.Key.ValueString()).
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *localResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state LocalResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var repo LocalAPIModel
	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("key", state.Key.ValueString()).
		SetResult(&repo).
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *localResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalResourceModel
	var state LocalResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repo LocalAPIModel
	resp.Diagnostics.Append(plan.ToAPIModel(ctx, r.PackageType, &repo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var jfrogErrors util.JFrogErrors
	response, err := r.ProviderData.Client.R().
		SetPathParam("key", plan.Key.ValueString()).
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

	if !plan.ProjectKey.Equal(state.ProjectKey) {
		key := plan.Key.ValueString()
		oldProjectKey := state.ProjectKey.ValueString()
		newProjectKey := plan.ProjectKey.ValueString()

		assignToProject := oldProjectKey == "" && len(newProjectKey) > 0
		unassignFromProject := len(oldProjectKey) > 0 && newProjectKey == ""

		var err error
		if assignToProject {
			err = repository.AssignRepoToProject(key, newProjectKey, r.ProviderData.Client)
		} else if unassignFromProject {
			err = repository.UnassignRepoFromProject(key, r.ProviderData.Client)
		}

		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to assign/unassign repository to project",
				err.Error(),
			)
			return
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *localResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state LocalResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var jfrogErrors util.JFrogErrors

	response, err := r.ProviderData.Client.R().
		SetPathParam("key", state.Key.ValueString()).
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

type LocalResourceModel struct {
	repository.BaseResourceModel
	BlackedOut             types.Bool `tfsdk:"blacked_out"`
	XrayIndex              types.Bool `tfsdk:"xray_index"`
	PropertySets           types.Set  `tfsdk:"property_sets"`
	ArchiveBrowsingEnabled types.Bool `tfsdk:"archive_browsing_enabled"`
	DownloadDirect         types.Bool `tfsdk:"download_direct"`
	PriorityResolution     types.Bool `tfsdk:"priority_resolution"`
}

func (r LocalResourceModel) ToAPIModel(ctx context.Context, packageType string, apiModel *LocalAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	var baseRepositoryAPIModel repository.BaseAPIModel

	r.BaseResourceModel.ToAPIModel(ctx, Rclass, packageType, &baseRepositoryAPIModel)

	var propertySets []string
	d := r.PropertySets.ElementsAs(ctx, &propertySets, false)
	if d != nil {
		diags.Append(d...)
	}

	*apiModel = LocalAPIModel{
		BaseAPIModel:           baseRepositoryAPIModel,
		BlackedOut:             r.BlackedOut.ValueBoolPointer(),
		XrayIndex:              r.XrayIndex.ValueBool(),
		PropertySets:           propertySets,
		ArchiveBrowsingEnabled: r.ArchiveBrowsingEnabled.ValueBoolPointer(),
		DownloadRedirect:       r.DownloadDirect.ValueBoolPointer(),
		PriorityResolution:     r.PriorityResolution.ValueBool(),
	}

	return diags
}

func (r *LocalResourceModel) FromAPIModel(ctx context.Context, apiModel LocalAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.BaseResourceModel.FromAPIModel(ctx, apiModel.BaseAPIModel)

	r.BlackedOut = types.BoolPointerValue(apiModel.BlackedOut)
	r.XrayIndex = types.BoolValue(apiModel.XrayIndex)
	r.ArchiveBrowsingEnabled = types.BoolPointerValue(apiModel.ArchiveBrowsingEnabled)
	r.DownloadDirect = types.BoolPointerValue(apiModel.DownloadRedirect)
	r.PriorityResolution = types.BoolValue(apiModel.PriorityResolution)

	propertySets, ds := types.SetValueFrom(ctx, types.StringType, apiModel.PropertySets)
	if ds.HasError() {
		diags.Append(ds...)
	}

	r.PropertySets = propertySets

	return diags
}

type LocalGenericResourceModel struct {
	LocalResourceModel
	RepoLayoutRef types.String `tfsdk:"repo_layout_ref"`
	CDNRedirect   types.Bool   `tfsdk:"cdn_redirect"`
}

func (r *LocalGenericResourceModel) fromAPIModel(ctx context.Context, apiModel LocalGenericAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.BaseResourceModel.FromAPIModel(ctx, apiModel.BaseAPIModel)
	r.LocalResourceModel.FromAPIModel(ctx, apiModel.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(apiModel.RepoLayoutRef)
	r.CDNRedirect = types.BoolPointerValue(apiModel.CDNRedirect)

	return diags
}

func (r LocalGenericResourceModel) ToAPIModel(ctx context.Context, packageType string, apiModel *LocalGenericAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	var localAPIModel LocalAPIModel
	r.LocalResourceModel.ToAPIModel(ctx, packageType, &localAPIModel)

	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	*apiModel = LocalGenericAPIModel{
		LocalAPIModel: localAPIModel,
		CDNRedirect:   r.CDNRedirect.ValueBoolPointer(),
	}

	return diags
}

type LocalAPIModel struct {
	repository.BaseAPIModel
	BlackedOut             *bool    `json:"blackedOut"`
	XrayIndex              bool     `json:"xrayIndex"`
	PropertySets           []string `json:"propertySets,omitempty"`
	ArchiveBrowsingEnabled *bool    `json:"archiveBrowsingEnabled"`
	DownloadRedirect       *bool    `json:"downloadRedirect"`
	PriorityResolution     bool     `json:"priorityResolution"`
}

type LocalGenericAPIModel struct {
	LocalAPIModel
	CDNRedirect *bool `json:"cdnRedirect"`
}

var LocalAttributes = lo.Assign(
	repository.BaseAttributes,
	map[string]schema.Attribute{
		"blacked_out": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, the repository does not participate in artifact resolution and new artifacts cannot be deployed.",
		},
		"xray_index": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.",
		},
		"priority_resolution": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
		},
		"property_sets": schema.SetAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Validators: []validator.Set{
				setvalidator.SizeAtLeast(1),
			},
			MarkdownDescription: "List of property set name",
		},
		"archive_browsing_enabled": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, you may view content such as HTML or Javadoc files directly from Artifactory.\nThis may not be safe and therefore requires strict content moderation to prevent malicious users from uploading content that may compromise security (e.g., cross-site scripting attacks).",
		},
		"download_direct": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.",
		},
	},
)

var LocalGenericAttributes = lo.Assign(
	LocalAttributes,
	map[string]schema.Attribute{
		"cdn_redirect": schema.BoolAttribute{
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is `false`",
		},
	},
)

type RepositoryBaseParams struct {
	Key                    string   `hcl:"key" json:"key,omitempty"`
	ProjectKey             string   `json:"projectKey"`
	ProjectEnvironments    []string `json:"environments"`
	Rclass                 string   `json:"rclass"`
	PackageType            string   `hcl:"package_type" json:"packageType,omitempty"`
	Description            string   `json:"description"`
	Notes                  string   `json:"notes"`
	IncludesPattern        string   `hcl:"includes_pattern" json:"includesPattern,omitempty"`
	ExcludesPattern        string   `hcl:"excludes_pattern" json:"excludesPattern,omitempty"`
	RepoLayoutRef          string   `hcl:"repo_layout_ref" json:"repoLayoutRef,omitempty"`
	BlackedOut             *bool    `hcl:"blacked_out" json:"blackedOut,omitempty"`
	XrayIndex              bool     `json:"xrayIndex"`
	PropertySets           []string `hcl:"property_sets" json:"propertySets,omitempty"`
	ArchiveBrowsingEnabled *bool    `hcl:"archive_browsing_enabled" json:"archiveBrowsingEnabled,omitempty"`
	DownloadRedirect       *bool    `hcl:"download_direct" json:"downloadRedirect,omitempty"`
	CdnRedirect            *bool    `json:"cdnRedirect"`
	PriorityResolution     bool     `hcl:"priority_resolution" json:"priorityResolution"`
	TerraformType          string   `json:"terraformType"`
}

func (bp RepositoryBaseParams) Id() string {
	return bp.Key
}

var baseSchema = map[string]*sdkv2_schema.Schema{
	"blacked_out": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, the repository does not participate in artifact resolution and new artifacts cannot be deployed.",
	},
	"xray_index": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable Indexing In Xray. Repository will be indexed with the default retention period. You will be able to change it via Xray settings.",
	},
	"priority_resolution": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Setting repositories with priority will cause metadata to be merged only from repositories set with this field",
	},
	"property_sets": {
		Type:        sdkv2_schema.TypeSet,
		Elem:        &sdkv2_schema.Schema{Type: sdkv2_schema.TypeString},
		Set:         sdkv2_schema.HashString,
		Optional:    true,
		Description: "List of property set name",
	},
	"archive_browsing_enabled": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Description: "When set, you may view content such as HTML or Javadoc files directly from Artifactory.\nThis may not be safe and therefore requires strict content moderation to prevent malicious users from uploading content that may compromise security (e.g., cross-site scripting attacks).",
	},
	"download_direct": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Description: "When set, download requests to this repository will redirect the client to download the artifact directly from the cloud storage provider. Available in Enterprise+ and Edge licenses only.",
	},
	"cdn_redirect": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is 'false'",
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

// GetPackageType `packageType` in the API call payload for Terraform repositories must be "terraform", but we use
// `terraform_module` and `terraform_provider` as a package types in the Provider. GetPackageType function corrects this discrepancy.
func GetPackageType(packageType string) string {
	if strings.Contains(packageType, "terraform_") {
		return "terraform"
	}
	return packageType
}

func UnpackBaseRepo(rclassType string, s *sdkv2_schema.ResourceData, packageType string) RepositoryBaseParams {
	d := &utilsdk.ResourceData{ResourceData: s}
	return RepositoryBaseParams{
		Rclass:                 rclassType,
		Key:                    d.GetString("key", false),
		ProjectKey:             d.GetString("project_key", false),
		ProjectEnvironments:    d.GetSet("project_environments"),
		PackageType:            GetPackageType(packageType),
		Description:            d.GetString("description", false),
		Notes:                  d.GetString("notes", false),
		IncludesPattern:        d.GetString("includes_pattern", false),
		ExcludesPattern:        d.GetString("excludes_pattern", false),
		RepoLayoutRef:          d.GetString("repo_layout_ref", false),
		BlackedOut:             d.GetBoolRef("blacked_out", false),
		ArchiveBrowsingEnabled: d.GetBoolRef("archive_browsing_enabled", false),
		PropertySets:           d.GetSet("property_sets"),
		XrayIndex:              d.GetBool("xray_index", false),
		DownloadRedirect:       d.GetBoolRef("download_direct", false),
		CdnRedirect:            d.GetBoolRef("cdn_redirect", false),
		PriorityResolution:     d.GetBool("priority_resolution", false),
	}
}
