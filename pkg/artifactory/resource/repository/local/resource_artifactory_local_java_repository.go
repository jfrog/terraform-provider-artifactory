package local

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	sdkv2_validator "github.com/jfrog/terraform-provider-shared/validator"
	"github.com/samber/lo"
)

func NewJavaLocalRepositoryResource(packageType string, suppressPom bool) func() resource.Resource {
	return func() resource.Resource {
		return &localJavaResource{
			localResource: NewLocalRepositoryResource(
				packageType,
				repository.PackageNameLookup[packageType],
				reflect.TypeFor[LocalJavaResourceModel](),
				reflect.TypeFor[LocalJavaAPIModel](),
			),
			suppressPom: suppressPom,
		}
	}
}

type localJavaResource struct {
	localResource
	suppressPom bool
}

type LocalJavaResourceModel struct {
	LocalResourceModel
	ChecksumPolicyType           types.String `tfsdk:"checksum_policy_type"`
	SnapshotVersionBehavior      types.String `tfsdk:"snapshot_version_behavior"`
	MaxUniqueSnapshots           types.Int64  `tfsdk:"max_unique_snapshots"`
	HandleReleases               types.Bool   `tfsdk:"handle_releases"`
	HandleSnapshots              types.Bool   `tfsdk:"handle_snapshots"`
	SuppressPOMConsistencyChecks types.Bool   `tfsdk:"suppress_pom_consistency_checks"`
}

func (r *LocalJavaResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r LocalJavaResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalJavaResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalJavaResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *LocalJavaResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *LocalJavaResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r LocalJavaResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r LocalJavaResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	model, d := r.LocalResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	localAPIModel := model.(LocalAPIModel)
	localAPIModel.RepoLayoutRef = r.RepoLayoutRef.ValueString()

	return LocalJavaAPIModel{
		LocalAPIModel:                localAPIModel,
		ChecksumPolicyType:           r.ChecksumPolicyType.ValueString(),
		SnapshotVersionBehavior:      r.SnapshotVersionBehavior.ValueString(),
		MaxUniqueSnapshots:           r.MaxUniqueSnapshots.ValueInt64(),
		HandleReleases:               r.HandleReleases.ValueBool(),
		HandleSnapshots:              r.HandleSnapshots.ValueBool(),
		SuppressPOMConsistencyChecks: r.SuppressPOMConsistencyChecks.ValueBool(),
	}, diags
}

func (r *LocalJavaResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*LocalJavaAPIModel)

	r.LocalResourceModel.FromAPIModel(ctx, model.LocalAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.ChecksumPolicyType = types.StringValue(model.ChecksumPolicyType)
	r.SnapshotVersionBehavior = types.StringValue(model.SnapshotVersionBehavior)
	r.MaxUniqueSnapshots = types.Int64Value(model.MaxUniqueSnapshots)
	r.HandleReleases = types.BoolValue(model.HandleReleases)
	r.HandleSnapshots = types.BoolValue(model.HandleSnapshots)
	r.SuppressPOMConsistencyChecks = types.BoolValue(model.SuppressPOMConsistencyChecks)

	return diags
}

type LocalJavaAPIModel struct {
	LocalAPIModel
	ChecksumPolicyType           string `json:"checksumPolicyType"`
	SnapshotVersionBehavior      string `json:"snapshotVersionBehavior"`
	MaxUniqueSnapshots           int64  `json:"maxUniqueSnapshots"`
	HandleReleases               bool   `json:"handleReleases"`
	HandleSnapshots              bool   `json:"handleSnapshots"`
	SuppressPOMConsistencyChecks bool   `json:"suppressPomConsistencyChecks"`
}

func (r *localJavaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(r.Rclass, r.PackageType),
		map[string]schema.Attribute{
			"checksum_policy_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("client-checksums"),
				Validators: []validator.String{
					stringvalidator.OneOf("client-checksums", "server-generated-checksums"),
				},
				MarkdownDescription: "Checksum policy determines how Artifactory behaves when a client checksum for a deployed " +
					"resource is missing or conflicts with the locally calculated checksum (bad checksum). " +
					`Options are: "client-checksums", or "server-generated-checksums". Default: "client-checksums"\n ` +
					"For more details, please refer to Checksum Policy - " +
					"https://www.jfrog.com/confluence/display/JFROG/Local+Repositories#LocalRepositories-ChecksumPolicy",
			},
			"snapshot_version_behavior": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("unique"),
				Validators: []validator.String{
					stringvalidator.OneOf("unique", "non-unique", "deployer"),
				},
				MarkdownDescription: "Specifies the naming convention for Maven SNAPSHOT versions. The options are - " +
					"`unique`: Version number is based on a time-stamp (default), " +
					"`non-unique`: Version number uses a self-overriding naming pattern of artifactId-version-SNAPSHOT.type, " +
					"`deployer`: Respects the settings in the Maven client that is deploying the artifact.",
			},
			"max_unique_snapshots": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
				MarkdownDescription: "The maximum number of unique snapshots of a single artifact to store. Once the number of " +
					"snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is " +
					"no limit, and unique snapshots are not cleaned up.",
			},
			"handle_releases": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "If set, Artifactory allows you to deploy release artifacts into this repository.",
			},
			"handle_snapshots": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "If set, Artifactory allows you to deploy snapshot artifacts into this repository.",
			},
			"suppress_pom_consistency_checks": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(r.suppressPom),
				MarkdownDescription: "By default, Artifactory keeps your repositories healthy by refusing POMs with incorrect " +
					"coordinates (path). If the groupId:artifactId:version information inside the POM does not match the " +
					"deployed path, Artifactory rejects the deployment with a `409 Conflict` error. You can disable this " +
					"behavior by setting the Suppress POM Consistency Checks checkbox.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  attributes,
		Description: r.Description,
	}
}

func GetJavaSchemas(packageType string, suppressPom bool) map[int16]map[string]*sdkv2_schema.Schema {
	javaSchema := lo.Assign(
		map[string]*sdkv2_schema.Schema{
			"checksum_policy_type": {
				Type:             sdkv2_schema.TypeString,
				Optional:         true,
				Default:          "client-checksums",
				ValidateDiagFunc: sdkv2_validator.StringInSlice(true, "client-checksums", "server-generated-checksums"),
				Description: "Checksum policy determines how Artifactory behaves when a client checksum for a deployed " +
					"resource is missing or conflicts with the locally calculated checksum (bad checksum). " +
					`Options are: "client-checksums", or "server-generated-checksums". Default: "client-checksums"\n ` +
					"For more details, please refer to Checksum Policy - " +
					"https://www.jfrog.com/confluence/display/JFROG/Local+Repositories#LocalRepositories-ChecksumPolicy",
			},
			"snapshot_version_behavior": {
				Type:             sdkv2_schema.TypeString,
				Optional:         true,
				Default:          "unique",
				ValidateDiagFunc: sdkv2_validator.StringInSlice(true, "unique", "non-unique", "deployer"),
				Description: "Specifies the naming convention for Maven SNAPSHOT versions.\nThe options are " +
					"-\nunique: Version number is based on a time-stamp (default)\nnon-unique: Version number uses a" +
					" self-overriding naming pattern of artifactId-version-SNAPSHOT.type\ndeployer: Respects the settings " +
					"in the Maven client that is deploying the artifact.",
			},
			"max_unique_snapshots": {
				Type:             sdkv2_schema.TypeInt,
				Optional:         true,
				Default:          0,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
				Description: "The maximum number of unique snapshots of a single artifact to store.\nOnce the number of " +
					"snapshots exceeds this setting, older versions are removed.\nA value of 0 (default) indicates there is " +
					"no limit, and unique snapshots are not cleaned up.",
			},
			"handle_releases": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If set, Artifactory allows you to deploy release artifacts into this repository.",
			},
			"handle_snapshots": {
				Type:        sdkv2_schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If set, Artifactory allows you to deploy snapshot artifacts into this repository.",
			},
			"suppress_pom_consistency_checks": {
				Type:     sdkv2_schema.TypeBool,
				Optional: true,
				Default:  suppressPom,
				Description: "By default, Artifactory keeps your repositories healthy by refusing POMs with incorrect " +
					"coordinates (path).\n  If the groupId:artifactId:version information inside the POM does not match the " +
					"deployed path, Artifactory rejects the deployment with a \"409 Conflict\" error.\n  You can disable this " +
					"behavior by setting the Suppress POM Consistency Checks checkbox.",
			},
		},
		repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
	)

	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(
			BaseSchemaV1,
			javaSchema,
		),
		1: lo.Assign(
			BaseSchemaV1,
			javaSchema,
		),
	}
}

type JavaLocalRepositoryParams struct {
	RepositoryBaseParams
	ChecksumPolicyType           string `hcl:"checksum_policy_type" json:"checksumPolicyType"`
	SnapshotVersionBehavior      string `hcl:"snapshot_version_behavior" json:"snapshotVersionBehavior"`
	MaxUniqueSnapshots           int    `hcl:"max_unique_snapshots" json:"maxUniqueSnapshots"`
	HandleReleases               bool   `hcl:"handle_releases" json:"handleReleases"`
	HandleSnapshots              bool   `hcl:"handle_snapshots" json:"handleSnapshots"`
	SuppressPomConsistencyChecks bool   `hcl:"suppress_pom_consistency_checks" json:"suppressPomConsistencyChecks"`
}
