package remote

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewVCSRemoteRepositoryResource() resource.Resource {
	return &remoteVCSResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.VCSPackageType,
			repository.PackageNameLookup[repository.VCSPackageType],
			reflect.TypeFor[remoteVCSResourceModel](),
			reflect.TypeFor[RemoteVCSAPIModel](),
		),
	}
}

type remoteVCSResource struct {
	remoteResource
}

type remoteVCSResourceModel struct {
	RemoteResourceModel
	vcsResourceModel
	MaxUniqueSnapshots types.Int64 `tfsdk:"max_unique_snapshots"`
}

func (r *remoteVCSResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read VCS plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteVCSResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into VCS state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteVCSResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read VCS state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteVCSResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into VCS state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteVCSResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read VCS state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteVCSResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read VCS state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteVCSResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into VCS state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteVCSResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteVCSAPIModel{
		RemoteAPIModel: remoteAPIModel,
		vcsAPIModel: vcsAPIModel{
			GitProvider:    r.VCSGitProvider.ValueStringPointer(),
			GitDownloadURL: r.VCSGitDownloadURL.ValueStringPointer(),
		},
		MaxUniqueSnapshots: r.MaxUniqueSnapshots.ValueInt64(),
	}, diags
}

func (r *remoteVCSResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteVCSAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.VCSGitProvider = types.StringPointerValue(model.vcsAPIModel.GitProvider)
	r.VCSGitDownloadURL = types.StringPointerValue(model.vcsAPIModel.GitDownloadURL)
	r.MaxUniqueSnapshots = types.Int64Value(model.MaxUniqueSnapshots)
	return diags
}

type RemoteVCSAPIModel struct {
	RemoteAPIModel
	vcsAPIModel
	MaxUniqueSnapshots int64 `json:"maxUniqueSnapshots"`
}

func (r *remoteVCSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteVCSAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		vcsAttributes,
		map[string]schema.Attribute{
			"max_unique_snapshots": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				MarkdownDescription: "The maximum number of unique snapshots of a single artifact to store. Once the number of " +
					"snapshots exceeds this setting, older versions are removed. A value of 0 (default) indicates there is " +
					"no limit, and unique snapshots are not cleaned up.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteVCSAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
