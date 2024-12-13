package local

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	"github.com/samber/lo"
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

func NewGenericLocalRepositoryResource(packageType string) func() resource.Resource {
	return func() resource.Resource {
		return &localGenericResource{
			localResource: localResource{
				BaseResource: repository.BaseResource{
					JFrogResource: util.JFrogResource{
						TypeName:           fmt.Sprintf("artifactory_local_%s_repository", packageType),
						CollectionEndpoint: "artifactory/api/repositories",
						DocumentEndpoint:   "artifactory/api/repositories/{key}",
					},
					Description: "Provides a resource to creates a local Machine Learning repository.",
					PackageType: packageType,
					Rclass:      Rclass,
				},
			},
		}
	}
}

type localGenericResource struct {
	localResource
}

type LocalGenericResourceModel struct {
	LocalResourceModel
	RepoLayoutRef types.String `tfsdk:"repo_layout_ref"`
	CDNRedirect   types.Bool   `tfsdk:"cdn_redirect"`
}

func (r *LocalGenericResourceModel) FromAPIModel(ctx context.Context, apiModel LocalGenericAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

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

type LocalGenericAPIModel struct {
	LocalAPIModel
	CDNRedirect *bool `json:"cdnRedirect"`
}

func (r *localGenericResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	localGenericAttributes := lo.Assign(
		LocalAttributes,
		repository.RepoLayoutRefAttribute(Rclass, r.PackageType),
		map[string]schema.Attribute{
			"cdn_redirect": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "When set, download requests to this repository will redirect the client to download the artifact directly from AWS CloudFront. Available in Enterprise+ and Edge licenses only. Default value is `false`",
			},
		},
	)

	resp.Schema = schema.Schema{
		Version:     1,
		Attributes:  localGenericAttributes,
		Description: r.Description,
	}
}

func (r *localGenericResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalGenericResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repo LocalGenericAPIModel
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

func (r *localGenericResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state LocalGenericResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert from Terraform data model into API data model
	var repo LocalGenericAPIModel
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

func (r *localGenericResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan LocalGenericResourceModel
	var state LocalGenericResourceModel

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

	var repo LocalGenericAPIModel
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

func GetGenericSchemas(packageType string) map[int16]map[string]*sdkv2_schema.Schema {
	return map[int16]map[string]*sdkv2_schema.Schema{
		0: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
		),
		1: lo.Assign(
			BaseSchemaV1,
			repository.RepoLayoutRefSDKv2Schema(Rclass, packageType),
		),
	}
}
