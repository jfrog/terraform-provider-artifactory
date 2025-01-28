package remote

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewGradleRemoteRepositoryResource() resource.Resource {
	return &remoteGradleResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.GradlePackageType,
			repository.PackageNameLookup[repository.GradlePackageType],
			reflect.TypeFor[remoteGradleResourceModel](),
			reflect.TypeFor[RemoteGradleAPIModel](),
		),
	}
}

type GradleRemoteRepo struct {
	RepositoryCurationParams
	JavaRemoteRepo
}

type remoteGradleResource struct {
	remoteResource
}

type remoteGradleResourceModel struct {
	RemoteResourceModel
	CurationResourceModel
	JavaResourceModel
}

func (r *remoteGradleResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteGradleResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGradleResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGradleResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteGradleResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteGradleResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteGradleResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteGradleResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteGradleAPIModel{
		RemoteAPIModel: remoteAPIModel,
		CurationAPIModel: CurationAPIModel{
			Curated: r.Curated.ValueBool(),
		},
		JavaAPIModel: JavaAPIModel{
			FetchJarsEagerly:             r.FetchJarsEagerly.ValueBool(),
			FetchSourcesEagerly:          r.FetchSourcesEagerly.ValueBool(),
			RemoteRepoChecksumPolicyType: r.RemoteRepoChecksumPolicyType.ValueString(),
			HandleReleases:               r.HandleReleases.ValueBool(),
			HandleSnapshots:              r.HandleSnapshots.ValueBool(),
			SuppressPomConsistencyChecks: r.SuppressPomConsistencyChecks.ValueBool(),
			RejectInvalidJars:            r.RejectInvalidJars.ValueBool(),
			MaxUniqueSnapshots:           r.MaxUniqueSnapshots.ValueInt64(),
		},
	}, diags
}

func (r *remoteGradleResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteGradleAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.Curated = types.BoolValue(model.CurationAPIModel.Curated)
	r.FetchJarsEagerly = types.BoolValue(model.JavaAPIModel.FetchJarsEagerly)
	r.FetchSourcesEagerly = types.BoolValue(model.JavaAPIModel.FetchSourcesEagerly)
	r.RemoteRepoChecksumPolicyType = types.StringValue(model.JavaAPIModel.RemoteRepoChecksumPolicyType)
	r.HandleReleases = types.BoolValue(model.JavaAPIModel.HandleReleases)
	r.HandleSnapshots = types.BoolValue(model.JavaAPIModel.HandleSnapshots)
	r.SuppressPomConsistencyChecks = types.BoolValue(model.JavaAPIModel.SuppressPomConsistencyChecks)
	r.RejectInvalidJars = types.BoolValue(model.JavaAPIModel.RejectInvalidJars)
	r.MaxUniqueSnapshots = types.Int64Value(model.JavaAPIModel.MaxUniqueSnapshots)

	return diags
}

type RemoteGradleAPIModel struct {
	RemoteAPIModel
	CurationAPIModel
	JavaAPIModel
}

func (r *remoteGradleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteGradleAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		CurationAttributes,
		javaAttributes(true),
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteGradleAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
