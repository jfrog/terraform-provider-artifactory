package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-shared/util"
	utilfw "github.com/jfrog/terraform-provider-shared/util/fw"

	"gopkg.in/yaml.v3"
)

func NewRepositoryLayoutResource() resource.Resource {
	return &RepositoryLayoutResource{
		TypeName: "artifactory_repository_layout",
	}
}

type RepositoryLayoutResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type RepositoryLayoutResourceModel struct {
	Name                             types.String `tfsdk:"name"`
	ArtifactPathPattern              types.String `tfsdk:"artifact_path_pattern"`
	DescriptorPathPattern            types.String `tfsdk:"descriptor_path_pattern"`
	DistinctiveDescriptorPathPattern types.Bool   `tfsdk:"distinctive_descriptor_path_pattern"`
	FileIntegrationRevisionRegExp    types.String `tfsdk:"file_integration_revision_regexp"`
	FolderIntegrationRevisionRegExp  types.String `tfsdk:"folder_integration_revision_regexp"`
}

type RepositoryLayoutAPIModel struct {
	Name                             string `xml:"name" yaml:"name"`
	ArtifactPathPattern              string `xml:"artifactPathPattern" yaml:"artifactPathPattern"`
	DescriptorPathPattern            string `xml:"descriptorPathPattern" yaml:"descriptorPathPattern"`
	DistinctiveDescriptorPathPattern bool   `xml:"distinctiveDescriptorPathPattern" yaml:"distinctiveDescriptorPathPattern"`
	FileIntegrationRevisionRegExp    string `xml:"fileIntegrationRevisionRegExp" yaml:"fileIntegrationRevisionRegExp"`
	FolderIntegrationRevisionRegExp  string `xml:"folderIntegrationRevisionRegExp" yaml:"folderIntegrationRevisionRegExp"`
}

func (m RepositoryLayoutAPIModel) Id() string {
	return m.Name
}

type RepositoryLayoutsAPIModel struct {
	Layouts []RepositoryLayoutAPIModel `xml:"repoLayouts>repoLayout" yaml:"repoLayout"`
}

func (r *RepositoryLayoutResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = r.TypeName
}

func (r *RepositoryLayoutResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Layout name",
			},
			"artifact_path_pattern": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "Please refer to: [Path Patterns](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts#RepositoryLayouts-ModulesandPathPatternsusedbyRepositoryLayouts) in the Artifactory Wiki documentation.",
			},
			"distinctive_descriptor_path_pattern": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),

				MarkdownDescription: "When set, `descriptor_path_pattern` will be used. Default to `false`.",
			},
			"descriptor_path_pattern": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Please refer to: [Descriptor Path Patterns](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts#RepositoryLayouts-DescriptorPathPatterns) in the Artifactory Wiki documentation",
			},
			"folder_integration_revision_regexp": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "A regular expression matching the integration revision string appearing in a folder name as part of the artifact's path. For example, `SNAPSHOT`, in Maven. Note! Take care not to introduce any regexp capturing groups within this expression. If not applicable use `.*`",
			},
			"file_integration_revision_regexp": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				MarkdownDescription: "A regular expression matching the integration revision string appearing in a file name as part of the artifact's path. For example, `SNAPSHOT|(?:(?:[0-9]{8}.[0-9]{6})-(?:[0-9]+))`, in Maven. Note! Take care not to introduce any regexp capturing groups within this expression. If not applicable use `.*`",
			},
		},
		MarkdownDescription: "Provides an Artifactory repository layout resource. See [Repository Layout documentation](https://www.jfrog.com/confluence/display/JFROG/Repository+Layouts) for more details.",
	}
}

func (r RepositoryLayoutResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data RepositoryLayoutResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.DistinctiveDescriptorPathPattern.ValueBool() && len(data.DescriptorPathPattern.ValueString()) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("descriptor_path_pattern"),
			"Invalid attribute configuration",
			"descriptor_path_pattern must be set when distinctive_descriptor_path_pattern is true",
		)
	}
}

func (r *RepositoryLayoutResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *RepositoryLayoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RepositoryLayoutResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryLayout := RepositoryLayoutAPIModel{
		Name:                             plan.Name.ValueString(),
		ArtifactPathPattern:              plan.ArtifactPathPattern.ValueString(),
		DescriptorPathPattern:            plan.DescriptorPathPattern.ValueString(),
		DistinctiveDescriptorPathPattern: plan.DistinctiveDescriptorPathPattern.ValueBool(),
		FileIntegrationRevisionRegExp:    plan.FileIntegrationRevisionRegExp.ValueString(),
		FolderIntegrationRevisionRegExp:  plan.FolderIntegrationRevisionRegExp.ValueString(),
	}

	///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	//GET call structure has "propertySets -> propertySet -> Array of property sets".
	//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
	//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
	//*/
	var body = map[string]map[string]interface{}{
		"repoLayouts": {
			repositoryLayout.Name: repositoryLayout,
		},
	}

	content, err := yaml.Marshal(&body)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("failed to marshal property set during Update: %s", err.Error()))
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToCreateResourceError(resp, fmt.Sprintf("failed to send PATCH request to Artifactory during Update: %s", err.Error()))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryLayoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepositoryLayoutResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repositoryLayouts RepositoryLayoutsAPIModel
	response, err := r.ProviderData.Client.R().
		SetResult(&repositoryLayouts).
		Get(ConfigurationEndpoint)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, fmt.Sprintf("failed to retrieve data from API: /artifactory/api/system/configuration during Read: %s", err.Error()))
		return
	}
	if response.IsError() {
		utilfw.UnableToRefreshResourceError(resp, fmt.Sprintf("failed to retrieve data from API: /artifactory/api/system/configuration during Read: %s", response.String()))
		return
	}

	matchedRepositoryLayout := FindConfigurationById(repositoryLayouts.Layouts, state.Name.ValueString())
	if matchedRepositoryLayout == nil {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("name"),
			"no matching repository layout found",
			state.Name.ValueString(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	state.Name = types.StringValue(matchedRepositoryLayout.Name)
	state.ArtifactPathPattern = types.StringValue(matchedRepositoryLayout.ArtifactPathPattern)
	state.DescriptorPathPattern = types.StringNull()
	if matchedRepositoryLayout.DistinctiveDescriptorPathPattern {
		state.DescriptorPathPattern = types.StringValue(matchedRepositoryLayout.DescriptorPathPattern)
	}
	state.DistinctiveDescriptorPathPattern = types.BoolValue(matchedRepositoryLayout.DistinctiveDescriptorPathPattern)
	state.FileIntegrationRevisionRegExp = types.StringValue(matchedRepositoryLayout.FileIntegrationRevisionRegExp)
	state.FolderIntegrationRevisionRegExp = types.StringValue(matchedRepositoryLayout.FolderIntegrationRevisionRegExp)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *RepositoryLayoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var plan RepositoryLayoutResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repositoryLayout := RepositoryLayoutAPIModel{
		Name:                             plan.Name.ValueString(),
		ArtifactPathPattern:              plan.ArtifactPathPattern.ValueString(),
		DescriptorPathPattern:            plan.DescriptorPathPattern.ValueString(),
		DistinctiveDescriptorPathPattern: plan.DistinctiveDescriptorPathPattern.ValueBool(),
		FileIntegrationRevisionRegExp:    plan.FileIntegrationRevisionRegExp.ValueString(),
		FolderIntegrationRevisionRegExp:  plan.FolderIntegrationRevisionRegExp.ValueString(),
	}

	///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	//GET call structure has "propertySets -> propertySet -> Array of property sets".
	//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
	//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
	//*/
	var body = map[string]map[string]interface{}{
		"repoLayouts": {
			repositoryLayout.Name: repositoryLayout,
		},
	}

	content, err := yaml.Marshal(&body)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("failed to marshal property set during Update: %s", err.Error()))
		return
	}

	err = SendConfigurationPatch(content, r.ProviderData)
	if err != nil {
		utilfw.UnableToUpdateResourceError(resp, fmt.Sprintf("failed to send PATCH request to Artifactory during Update: %s", err.Error()))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RepositoryLayoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client.R(), r.ProviderData.ProductId, r.TypeName)

	var state RepositoryLayoutResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var repoLayouts RepositoryLayoutsAPIModel
	response, err := r.ProviderData.Client.R().
		SetResult(&repoLayouts).
		Get(ConfigurationEndpoint)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, fmt.Sprintf("failed to retrieve data from API: /artifactory/api/system/configuration during Read: %s", err.Error()))
		return
	}
	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, fmt.Sprintf("got error response for API: /artifactory/api/system/configuration request during Read: %s", response.String()))
		return
	}

	matchedRepoLayout := FindConfigurationById(repoLayouts.Layouts, state.Name.ValueString())
	if matchedRepoLayout == nil {
		utilfw.UnableToDeleteResourceError(resp, fmt.Sprintf("No property set found for '%s'", state.Name.ValueString()))
		return
	}

	deleteConfig := fmt.Sprintf(`
repoLayouts:
  %s: ~
`, matchedRepoLayout.Name)

	err = SendConfigurationPatch([]byte(deleteConfig), r.ProviderData)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *RepositoryLayoutResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
