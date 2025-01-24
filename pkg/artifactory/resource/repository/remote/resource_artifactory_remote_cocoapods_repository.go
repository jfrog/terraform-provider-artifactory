package remote

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/samber/lo"
)

func NewCocoapodsRemoteRepositoryResource() resource.Resource {
	return &remoteCocoapodsResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.CocoapodsPackageType,
			repository.PackageNameLookup[repository.CocoapodsPackageType],
			reflect.TypeFor[remoteCocoapodsResourceModel](),
			reflect.TypeFor[RemoteCocoapodsAPIModel](),
		),
	}
}

type remoteCocoapodsResource struct {
	remoteResource
}

type remoteCocoapodsResourceModel struct {
	RemoteResourceModel
	vcsResourceModel
	PodsSpecsRepoURL types.String `tfsdk:"pods_specs_repo_url"`
}

func (r *remoteCocoapodsResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteCocoapodsResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteCocoapodsResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteCocoapodsResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteCocoapodsResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteCocoapodsResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteCocoapodsResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteCocoapodsResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteCocoapodsAPIModel{
		RemoteAPIModel: remoteAPIModel,
		vcsAPIModel: vcsAPIModel{
			GitProvider:    r.VCSGitProvider.ValueStringPointer(),
			GitDownloadURL: r.VCSGitDownloadURL.ValueStringPointer(),
		},
		PodsSpecsRepoURL: r.PodsSpecsRepoURL.ValueString(),
	}, diags
}

func (r *remoteCocoapodsResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteCocoapodsAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.VCSGitProvider = types.StringPointerValue(model.vcsAPIModel.GitProvider)
	r.VCSGitDownloadURL = types.StringPointerValue(model.vcsAPIModel.GitDownloadURL)
	r.PodsSpecsRepoURL = types.StringValue(model.PodsSpecsRepoURL)

	return diags
}

type RemoteCocoapodsAPIModel struct {
	RemoteAPIModel
	vcsAPIModel
	PodsSpecsRepoURL string `json:"podsSpecsRepoUrl"`
}

func (r *remoteCocoapodsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteCocoapodsAttributes := lo.Assign(
		RemoteAttributes,
		vcsAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"pods_specs_repo_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("https://github.com/CocoaPods/Specs"),
				MarkdownDescription: "Proxy remote CocoaPods Specs repositories. Default value is 'https://github.com/CocoaPods/Specs'.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteCocoapodsAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
