package artifact

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fw_path "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"
	validatorfw_string "github.com/jfrog/terraform-provider-shared/validator/fw/string"
	"github.com/samber/lo"
)

const itemPropertiesEndpoint = "/artifactory/api/storage/{repo_key}"

func NewItemPropertiesResource() resource.Resource {
	return &ItemPropertiesResource{
		TypeName: "artifactory_item_properties",
	}
}

type ItemPropertiesResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type ItemPropertiesResourceModel struct {
	RepoKey     types.String `tfsdk:"repo_key"`
	ItemPath    types.String `tfsdk:"item_path"`
	Properties  types.Map    `tfsdk:"properties"`
	IsRecursive types.Bool   `tfsdk:"is_recursive"`
}

func (r *ItemPropertiesResourceModel) toPropertiesQueryParamsString(ctx context.Context, params *string) diag.Diagnostics {
	// Convert from Terraform resource model into API model
	var properties map[string][]string
	diags := r.Properties.ElementsAs(ctx, &properties, false)
	if diags.HasError() {
		return diags
	}

	*params = lo.Reduce(
		lo.Keys(properties),
		func(val, key string, _ int) string {
			values := strings.Join(properties[key], ",")

			if val == "" {
				return fmt.Sprintf("%s=%s", key, values)
			}

			return fmt.Sprintf("%s;%s=%s", val, key, values)
		},
		"",
	)

	return nil
}

func (r *ItemPropertiesResourceModel) fromAPIModel(ctx context.Context, apiModel ItemPropertiesGetAPIModel) (ds diag.Diagnostics) {
	attrValues := lo.MapEntries(
		apiModel.Properties,
		func(k string, v []string) (string, attr.Value) {
			valueSet, d := types.SetValueFrom(ctx, types.StringType, v)
			if d.HasError() {
				ds.Append(d...)
			}

			return k, valueSet
		},
	)

	propertiesSet, d := types.MapValue(
		types.SetType{ElemType: types.StringType},
		attrValues,
	)
	if d.HasError() {
		ds.Append(d...)
	}

	r.Properties = propertiesSet

	return nil
}

func (r ItemPropertiesResourceModel) GetAPIRequestAndURL(req *resty.Request, baseURL string) (request *resty.Request, url string) {
	request = req.SetPathParam("repo_key", r.RepoKey.ValueString())

	url = baseURL

	if !r.ItemPath.IsNull() {
		url = path.Join(url, "{item_path}")

		request = req.SetRawPathParam("item_path", r.ItemPath.ValueString())
	}

	return
}

type ItemPropertiesGetAPIModel struct {
	URI        string              `json:"uri"`
	Properties map[string][]string `json:"properties"`
}

type ItemPropertiesPatchAPIModel struct {
	Props map[string]*string `json:"props"`
}

func (r *ItemPropertiesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *ItemPropertiesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repo_key": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					validatorfw_string.RepoKey(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "Respository key.",
			},
			"item_path": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				MarkdownDescription: "The relative path of the item (file/folder/repository). Leave unset for repository.",
			},
			"properties": schema.MapAttribute{
				ElementType: types.SetType{ElemType: types.StringType},
				Required:    true,
				Validators: []validator.Map{
					mapvalidator.SizeAtLeast(1),
					mapvalidator.KeysAre(
						stringvalidator.LengthBetween(1, 255),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[a-zA-Z].*`), "must begin with a letter"),
						validatorfw_string.RegexNotMatches(regexp.MustCompile(`[)(}{\]\[\-*+^$/~\x60!@#%&<>;=,±§\s]+`), "must not contain the following special characters: )(}{][-*+^$\\/~`!@#%&<>;=,±§ and the space character"),
					),
					mapvalidator.ValueSetsAre(
						setvalidator.SizeAtLeast(1),
						setvalidator.ValueStringsAre(
							stringvalidator.LengthBetween(1, 2400),
						),
					),
				},
				MarkdownDescription: "Map of key and list of values.\n\n~>Keys are limited up to 255 characters and values are limited up to 2,400 characters. Using properties with values over this limit might cause backend issues.\n\n" +
					"~>The following special characters are forbidden in the key field: `)(}{][*+^$/~``!@#%&<>;=,±§` and the space character.",
			},
			"is_recursive": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Add this property to the selected folder and to all of artifacts and folders under this folder. Default to `false`",
			},
		},
		MarkdownDescription: "Provides a resource for managaing item (file, folder, or repository) properties. When a folder is used property attachment is recursive by default. See [JFrog documentation](https://jfrog.com/help/r/jfrog-artifactory-documentation/working-with-jfrog-properties) for more details.",
	}
}

func (r *ItemPropertiesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *ItemPropertiesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ItemPropertiesResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var propertiesString string
	resp.Diagnostics.Append(plan.toPropertiesQueryParamsString(ctx, &propertiesString)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isRecursive := 0
	if plan.IsRecursive.ValueBool() {
		isRecursive = 1
	}

	request, url := plan.GetAPIRequestAndURL(r.ProviderData.Client.R(), itemPropertiesEndpoint)

	response, err := request.
		SetQueryParams(map[string]string{
			"properties": propertiesString,
			"recursive":  fmt.Sprintf("%d", isRecursive),
		}).
		Put(url)

	if err != nil {
		utilfw.UnableToCreateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToCreateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemPropertiesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ItemPropertiesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request, url := state.GetAPIRequestAndURL(r.ProviderData.Client.R(), itemPropertiesEndpoint)

	// Convert from Terraform data model into API data model
	var properties ItemPropertiesGetAPIModel
	response, err := request.
		SetQueryParam("properties", "").
		SetResult(&properties).
		Get(url)

	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, response.String())
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.fromAPIModel(ctx, properties)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ItemPropertiesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceUpdate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan ItemPropertiesResourceModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ItemPropertiesResourceModel
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var planProperties map[string][]string
	resp.Diagnostics.Append(plan.Properties.ElementsAs(ctx, &planProperties, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateProperties map[string][]string
	resp.Diagnostics.Append(state.Properties.ElementsAs(ctx, &stateProperties, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, propKeysToRemove := lo.Difference(
		lo.Keys(planProperties),
		lo.Keys(stateProperties),
	)

	props := lo.MapEntries(
		planProperties,
		func(k string, v []string) (string, *string) {
			str := strings.Join(v, ",")
			return k, &str
		},
	)

	for _, key := range propKeysToRemove {
		props[key] = nil
	}

	updateProps := ItemPropertiesPatchAPIModel{
		Props: props,
	}

	isRecursive := 0
	if plan.IsRecursive.ValueBool() {
		isRecursive = 1
	}

	request, url := plan.GetAPIRequestAndURL(r.ProviderData.Client.R(), "artifactory/api/metadata/{repo_key}")

	response, err := request.
		SetQueryParam("recursiveProperties", fmt.Sprintf("%d", isRecursive)).
		SetBody(updateProps).
		Patch(url)

	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToUpdateResourceError(resp, response.String())
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ItemPropertiesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state ItemPropertiesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	var properties map[string][]string
	resp.Diagnostics.Append(state.Properties.ElementsAs(ctx, &properties, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	isRecursive := 0
	if state.IsRecursive.ValueBool() {
		isRecursive = 1
	}

	request, url := state.GetAPIRequestAndURL(r.ProviderData.Client.R(), itemPropertiesEndpoint)

	response, err := request.
		SetQueryParams(map[string]string{
			"properties": strings.Join(lo.Keys(properties), ","),
			"recursive":  fmt.Sprintf("%d", isRecursive),
		}).
		Delete(url)

	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, response.String())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *ItemPropertiesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, fw_path.Root("repo_key"), parts[0])...)

	if len(parts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, fw_path.Root("item_path"), parts[1])...)
	}
}
