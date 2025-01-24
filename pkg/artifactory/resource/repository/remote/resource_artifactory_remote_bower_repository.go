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

func NewBowerRemoteRepositoryResource() resource.Resource {
	return &remoteBowerResource{
		remoteResource: NewRemoteRepositoryResource(
			repository.BowerPackageType,
			repository.PackageNameLookup[repository.BowerPackageType],
			reflect.TypeFor[remoteBowerResourceModel](),
			reflect.TypeFor[RemoteBowerAPIModel](),
		),
	}
}

type remoteBowerResource struct {
	remoteResource
}

type remoteBowerResourceModel struct {
	RemoteResourceModel
	vcsResourceModel
	BowerRegistryURL types.String `tfsdk:"bower_registry_url"`
}

func (r *remoteBowerResourceModel) GetCreateResourcePlanData(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r remoteBowerResourceModel) SetCreateResourceStateData(ctx context.Context, resp *resource.CreateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteBowerResourceModel) GetReadResourceStateData(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteBowerResourceModel) SetReadResourceStateData(ctx context.Context, resp *resource.ReadResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r *remoteBowerResourceModel) GetUpdateResourcePlanData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, r)...)
}

func (r *remoteBowerResourceModel) GetUpdateResourceStateData(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, r)...)
}

func (r remoteBowerResourceModel) SetUpdateResourceStateData(ctx context.Context, resp *resource.UpdateResponse) {
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &r)...)
}

func (r remoteBowerResourceModel) ToAPIModel(ctx context.Context, packageType string) (interface{}, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	remoteAPIModel, d := r.RemoteResourceModel.ToAPIModel(ctx, packageType)
	if d != nil {
		diags.Append(d...)
	}

	return RemoteBowerAPIModel{
		RemoteAPIModel: remoteAPIModel,
		vcsAPIModel: vcsAPIModel{
			GitProvider:    r.VCSGitProvider.ValueStringPointer(),
			GitDownloadURL: r.VCSGitDownloadURL.ValueStringPointer(),
		},
		BowerRegistryURL: r.BowerRegistryURL.ValueString(),
	}, diags
}

func (r *remoteBowerResourceModel) FromAPIModel(ctx context.Context, apiModel interface{}) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model := apiModel.(*RemoteBowerAPIModel)

	r.RemoteResourceModel.FromAPIModel(ctx, model.RemoteAPIModel)

	r.RepoLayoutRef = types.StringValue(model.RepoLayoutRef)
	r.BowerRegistryURL = types.StringValue(model.BowerRegistryURL)
	r.VCSGitProvider = types.StringPointerValue(model.vcsAPIModel.GitProvider)
	r.VCSGitDownloadURL = types.StringPointerValue(model.vcsAPIModel.GitDownloadURL)

	return diags
}

type RemoteBowerAPIModel struct {
	RemoteAPIModel
	vcsAPIModel
	BowerRegistryURL string `json:"bowerRegistryUrl"`
}

func (r *remoteBowerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	remoteBowerAttributes := lo.Assign(
		RemoteAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		vcsAttributes,
		map[string]schema.Attribute{
			"bower_registry_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("https://registry.bower.io"),
				MarkdownDescription: "Proxy remote Bower repository. Default value is 'https://registry.bower.io'.",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     CurrentSchemaVersion,
		Attributes:  remoteBowerAttributes,
		Blocks:      remoteBlocks,
		Description: r.Description,
	}
}
