package local

import (
	"context"
	"reflect"
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
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/samber/lo"
)

const (
	Rclass               = "local"
	CurrentSchemaVersion = 1
)

func NewLocalRepositoryResource(packageType, packageName string, resourceModelType, apiModelType reflect.Type) localResource {
	return localResource{
		BaseResource: repository.NewRepositoryResource(packageType, packageName, Rclass, resourceModelType, apiModelType),
	}
}

type localResource struct {
	repository.BaseResource
}

type LocalResourceModel struct {
	repository.BaseResourceModel
	BlackedOut             types.Bool   `tfsdk:"blacked_out"`
	XrayIndex              types.Bool   `tfsdk:"xray_index"`
	PropertySets           types.Set    `tfsdk:"property_sets"`
	ArchiveBrowsingEnabled types.Bool   `tfsdk:"archive_browsing_enabled"`
	DownloadDirect         types.Bool   `tfsdk:"download_direct"`
	PriorityResolution     types.Bool   `tfsdk:"priority_resolution"`
	RepoLayoutRef          types.String `tfsdk:"repo_layout_ref"`
	CDNRedirect            types.Bool   `tfsdk:"cdn_redirect"`
}

func (r *LocalResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.BaseResourceModel.ToAPIModel(ctx, Rclass, packageType)
	if d != nil {
		diags.Append(d...)
	}
	baseRepositoryAPIModel := model.(repository.BaseAPIModel)

	if r.RepoLayoutRef.IsNull() {
		repoLayoutRef, err := repository.GetDefaultRepoLayoutRef(Rclass, packageType)
		if err != nil {
			diags.AddError(
				"Failed to get default repo layout ref",
				err.Error(),
			)
		}
		baseRepositoryAPIModel.RepoLayoutRef = repoLayoutRef
	} else {
		baseRepositoryAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()
	}

	var propertySets []string
	d = r.PropertySets.ElementsAs(ctx, &propertySets, false)
	if d != nil {
		diags.Append(d...)
	}

	return LocalAPIModel{
		BaseAPIModel:           baseRepositoryAPIModel,
		BlackedOut:             r.BlackedOut.ValueBool(),
		XrayIndex:              r.XrayIndex.ValueBool(),
		PropertySets:           propertySets,
		ArchiveBrowsingEnabled: r.ArchiveBrowsingEnabled.ValueBool(),
		DownloadRedirect:       r.DownloadDirect.ValueBool(),
		PriorityResolution:     r.PriorityResolution.ValueBool(),
		CDNRedirect:            r.CDNRedirect.ValueBool(),
	}, diags
}

func (r *LocalResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(LocalAPIModel)

	r.BaseResourceModel.FromAPIModel(ctx, model.BaseAPIModel)

	r.BlackedOut = types.BoolValue(model.BlackedOut)
	r.XrayIndex = types.BoolValue(model.XrayIndex)
	r.ArchiveBrowsingEnabled = types.BoolValue(model.ArchiveBrowsingEnabled)
	r.DownloadDirect = types.BoolValue(model.DownloadRedirect)
	r.PriorityResolution = types.BoolValue(model.PriorityResolution)
	r.CDNRedirect = types.BoolValue(model.CDNRedirect)

	var propertySets = types.SetNull(types.StringType)
	if len(model.PropertySets) > 0 {
		ps, ds := types.SetValueFrom(ctx, types.StringType, model.PropertySets)
		if ds.HasError() {
			diags.Append(ds...)
			return diags
		}

		propertySets = ps
	}

	r.PropertySets = propertySets

	return diags
}

type LocalAPIModel struct {
	repository.BaseAPIModel
	BlackedOut             bool     `json:"blackedOut"`
	XrayIndex              bool     `json:"xrayIndex"`
	PropertySets           []string `json:"propertySets,omitempty"`
	ArchiveBrowsingEnabled bool     `json:"archiveBrowsingEnabled"`
	DownloadRedirect       bool     `json:"downloadRedirect"`
	PriorityResolution     bool     `json:"priorityResolution"`
	CDNRedirect            bool     `json:"cdnRedirect"`
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
		"cdn_redirect": schema.BoolAttribute{ // For backward compatibility with SDKv2. Only generic repo uses this
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is 'false'",
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
