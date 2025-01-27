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

func NewJavaRemoteRepositoryResource(packageType string, suppressPOM bool) func() resource.Resource {
	return func() resource.Resource {
		return &remoteJavaResource{
			remoteResource: NewRemoteRepositoryResource(
				packageType,
				repository.PackageNameLookup[packageType],
				reflect.TypeFor[remoteJavaResourceModel](),
				reflect.TypeFor[RemoteJavaAPIModel](),
			),
			suppressPOM: suppressPOM,
		}
	}
}

type remoteJavaResource struct {
	remoteResource
	suppressPOM bool
}

type remoteJavaResourceModel struct {
	RemoteResourceModel
	JavaResourceModel
}

func (r *remoteJavaResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteJavaResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteJavaResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteJavaResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteJavaResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteJavaResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteJavaResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteJavaResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteJavaAPIModel{
		RemoteAPIModel: remoteAPIModel,
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

func (r *remoteJavaResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteJavaAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
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

type RemoteJavaAPIModel struct {
	RemoteAPIModel
	JavaAPIModel
}

func (r *remoteJavaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteJavaAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		javaAttributes(r.suppressPOM),
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteJavaAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
