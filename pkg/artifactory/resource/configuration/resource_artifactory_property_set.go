package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func NewPropertySetResource() resource.Resource {
	return &PropertySetResource{}
}

type PropertySetResource struct {
	ProviderData util.ProviderMetadata
	TypeName     string
}

type PropertySetResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Visible  types.Bool   `tfsdk:"visible"`
	Property types.Set    `tfsdk:"property"`
}

func (r *PropertySetResourceModel) toAPIModel(_ context.Context, apiModel *PropertySetAPIModel) diag.Diagnostics {
	apiModel.Name = r.Name.ValueString()
	apiModel.Visible = r.Visible.ValueBool()

	properties := lo.Map(
		r.Property.Elements(),
		func(elem attr.Value, _ int) PropertyAPIModel {
			attrs := elem.(types.Object).Attributes()

			pvElems := attrs["predefined_value"].(types.Set).Elements()

			predefinedValues := lo.Map(
				pvElems,
				func(elem attr.Value, _ int) PredefinedValueAPIModel {
					attrs := elem.(types.Object).Attributes()

					return PredefinedValueAPIModel{
						Name:         attrs["name"].(types.String).ValueString(),
						DefaultValue: attrs["default_value"].(types.Bool).ValueBool(),
					}
				},
			)

			return PropertyAPIModel{
				Name:                  attrs["name"].(types.String).ValueString(),
				ClosedPredefinedValue: attrs["closed_predefined_values"].(types.Bool).ValueBool(),
				MultipleChoice:        attrs["multiple_choice"].(types.Bool).ValueBool(),
				PredefinedValues:      predefinedValues,
			}
		},
	)

	apiModel.Properties = properties

	return nil
}

var predefinedValueResourceModelAttributeTypes map[string]attr.Type = map[string]attr.Type{
	"name":          types.StringType,
	"default_value": types.BoolType,
}

var propertyResourceModelAttributeTypes map[string]attr.Type = map[string]attr.Type{
	"name":                     types.StringType,
	"closed_predefined_values": types.BoolType,
	"multiple_choice":          types.BoolType,
	"predefined_value": types.SetType{
		ElemType: types.ObjectType{
			AttrTypes: predefinedValueResourceModelAttributeTypes,
		},
	},
}

var propertySetResourceModelAttributeTypes types.ObjectType = types.ObjectType{
	AttrTypes: propertyResourceModelAttributeTypes,
}

func (r *PropertySetResourceModel) fromAPIModel(_ context.Context, apiModel PropertySetAPIModel) diag.Diagnostics {
	diags := diag.Diagnostics{}

	r.ID = types.StringValue(apiModel.Id())
	r.Name = types.StringValue(apiModel.Name)
	r.Visible = types.BoolValue(apiModel.Visible)

	property := lo.Map(
		apiModel.Properties,
		func(property PropertyAPIModel, _ int) attr.Value {
			predefinedValues := lo.Map(
				property.PredefinedValues,
				func(pv PredefinedValueAPIModel, _ int) attr.Value {
					p, ds := types.ObjectValue(
						predefinedValueResourceModelAttributeTypes,
						map[string]attr.Value{
							"name":          types.StringValue(pv.Name),
							"default_value": types.BoolValue(pv.DefaultValue),
						},
					)

					if ds != nil {
						diags = append(diags, ds...)
					}

					return p
				},
			)

			predefinedValue, ds := types.SetValue(
				types.ObjectType{
					AttrTypes: predefinedValueResourceModelAttributeTypes,
				},
				predefinedValues,
			)

			if ds != nil {
				diags = append(diags, ds...)
			}

			p, ds := types.ObjectValue(
				propertyResourceModelAttributeTypes,
				map[string]attr.Value{
					"name":                     types.StringValue(property.Name),
					"closed_predefined_values": types.BoolValue(property.ClosedPredefinedValue),
					"multiple_choice":          types.BoolValue(property.MultipleChoice),
					"predefined_value":         predefinedValue,
				},
			)

			if ds != nil {
				diags = append(diags, ds...)
			}

			return p
		},
	)

	p, ds := types.SetValue(
		propertySetResourceModelAttributeTypes,
		property,
	)

	if ds != nil {
		diags = append(diags, ds...)
	}

	r.Property = p

	return diags
}

type PredefinedValueAPIModel struct {
	Name         string `xml:"value" yaml:"-"`
	DefaultValue bool   `xml:"defaultValue" yaml:"defaultValue"`
}

type PropertyAPIModel struct {
	Name                  string                    `xml:"name" yaml:"-"`
	PredefinedValues      []PredefinedValueAPIModel `xml:"predefinedValues>predefinedValue" yaml:"predefinedValues"`
	ClosedPredefinedValue bool                      `xml:"closedPredefinedValues" yaml:"closedPredefinedValues"`
	MultipleChoice        bool                      `xml:"multipleChoice" yaml:"multipleChoice"`
}

type PropertySetAPIModel struct {
	Name       string             `xml:"name" yaml:"-"`
	Visible    bool               `xml:"visible" yaml:"visible"`
	Properties []PropertyAPIModel `xml:"properties>property" yaml:"properties"`
}

func (m PropertySetAPIModel) Id() string {
	return m.Name
}

type PropertySetsAPIModel struct {
	PropertySets []PropertySetAPIModel `xml:"propertySets>propertySet" yaml:"propertySet"`
}

func (r *PropertySetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_property_set"
	r.TypeName = resp.TypeName
}

func (r *PropertySetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Property set name.",
			},
			"visible": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Defines if the list visible and assignable to the repository or artifact.",
			},
		},
		Blocks: map[string]schema.Block{
			"property": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
							Description: "The name of the property.",
						},
						"closed_predefined_values": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Disables `multiple_choice` if set to `false` at the same time with `multiple_choice` set to `true`.",
						},
						"multiple_choice": schema.BoolAttribute{
							Optional:            true,
							Computed:            true,
							Default:             booldefault.StaticBool(false),
							MarkdownDescription: "Whether or not user can select multiple values. `closed_predefined_values` should be set to `true`.",
						},
					},
					Blocks: map[string]schema.Block{
						"predefined_value": schema.SetNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtLeast(1),
										},
										Description: "Predefined property name.",
									},
									"default_value": schema.BoolAttribute{
										Required:            true,
										MarkdownDescription: "Whether the value is selected by default in the UI.",
									},
								},
							},
							Validators: []validator.Set{
								setvalidator.IsRequired(),
							},
							Description: "Properties in the property set.",
						},
					},
					Validators: []validator.Object{
						propertyValidator{},
					},
				},
				Validators: []validator.Set{
					setvalidator.IsRequired(),
				},
				Description: "A list of properties that will be part of the property set.",
			},
		},
		Description: "Provides an Artifactory Property Set resource. This resource configuration corresponds to `propertySets` config block in system configuration XML (REST endpoint: artifactory/api/system/configuration).",
	}
}

func (r *PropertySetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	r.ProviderData = req.ProviderData.(util.ProviderMetadata)
}

func (r *PropertySetResource) transformPredefinedValues(values []PredefinedValueAPIModel) map[string]interface{} {
	transformedPredefinedValues := map[string]interface{}{}
	for _, value := range values {
		transformedPredefinedValues[value.Name] = map[string]interface{}{
			"defaultValue": value.DefaultValue,
		}
	}
	return transformedPredefinedValues
}

func (r *PropertySetResource) transformProperties(properties []PropertyAPIModel) map[string]interface{} {
	transformedProperties := map[string]interface{}{}

	for _, property := range properties {
		transformedProperties[property.Name] = map[string]interface{}{
			"predefinedValues":       r.transformPredefinedValues(property.PredefinedValues),
			"closedPredefinedValues": property.ClosedPredefinedValue,
			"multipleChoice":         property.MultipleChoice,
		}
	}
	return transformedProperties
}

func (r *PropertySetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var plan PropertySetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var propertySet PropertySetAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &propertySet)...)
	if resp.Diagnostics.HasError() {
		return
	}

	///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	//GET call structure has "propertySets -> propertySet -> Array of property sets".
	//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
	//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
	//*/
	var body = map[string]map[string]map[string]interface{}{
		"propertySets": {
			propertySet.Name: {
				"visible":    propertySet.Visible,
				"properties": r.transformProperties(propertySet.Properties),
			},
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

	plan.ID = types.StringValue(propertySet.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PropertySetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	go util.SendUsageResourceRead(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var state PropertySetResourceModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var propertySets PropertySetsAPIModel
	_, err := r.ProviderData.Client.R().
		SetResult(&propertySets).
		Get(ConfigurationEndpoint)
	if err != nil {
		utilfw.UnableToRefreshResourceError(resp, fmt.Sprintf("failed to retrieve data from API: /artifactory/api/system/configuration during Read: %s", err.Error()))
		return
	}

	matchedPropertySet := FindConfigurationById[PropertySetAPIModel](propertySets.PropertySets, state.Name.ValueString())
	if matchedPropertySet == nil {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("name"),
			"no matching property set found",
			state.Name.ValueString(),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Convert from the API data model to the Terraform data model
	// and refresh any attribute values.
	resp.Diagnostics.Append(state.fromAPIModel(ctx, *matchedPropertySet)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PropertySetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	go util.SendUsageResourceCreate(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var plan PropertySetResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var propertySet PropertySetAPIModel
	resp.Diagnostics.Append(plan.toAPIModel(ctx, &propertySet)...)
	if resp.Diagnostics.HasError() {
		return
	}

	///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
	//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
	//GET call structure has "propertySets -> propertySet -> Array of property sets".
	//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
	//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
	//*/
	var body = map[string]map[string]map[string]interface{}{
		"propertySets": {
			propertySet.Name: {
				"visible":    propertySet.Visible,
				"properties": r.transformProperties(propertySet.Properties),
			},
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

	plan.ID = types.StringValue(propertySet.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PropertySetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	go util.SendUsageResourceDelete(ctx, r.ProviderData.Client, r.ProviderData.ProductId, r.TypeName)

	var state PropertySetResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var propertySets PropertySetsAPIModel
	response, err := r.ProviderData.Client.R().
		SetResult(&propertySets).
		Get(ConfigurationEndpoint)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, fmt.Sprintf("failed to retrieve data from API: /artifactory/api/system/configuration during Read: %s", err.Error()))
		return
	}
	if response.IsError() {
		utilfw.UnableToDeleteResourceError(resp, fmt.Sprintf("got error response for API: /artifactory/api/system/configuration request during Read: %s", err.Error()))
		return
	}

	matchedPropertySet := FindConfigurationById[PropertySetAPIModel](propertySets.PropertySets, state.Name.ValueString())
	if matchedPropertySet == nil {
		utilfw.UnableToDeleteResourceError(resp, fmt.Sprintf("No property set found for '%s'", state.Name.ValueString()))
		return
	}

	deleteConfig := fmt.Sprintf(`
propertySets:
  %s: ~
`, matchedPropertySet.Name)

	err = SendConfigurationPatch([]byte(deleteConfig), r.ProviderData)
	if err != nil {
		utilfw.UnableToDeleteResourceError(resp, err.Error())
		return
	}

	// If the logic reaches here, it implicitly succeeded and will remove
	// the resource from state if there are no other errors.
}

// ImportState imports the resource into the Terraform state.
func (r *PropertySetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

type propertyValidator struct{}

func (v propertyValidator) Description(ctx context.Context) string {
	return ""
}

func (v propertyValidator) MarkdownDescription(ctx context.Context) string {
	return ""
}

func (v propertyValidator) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	attrs := req.ConfigValue.Attributes()

	closedPredefinedValues := attrs["closed_predefined_values"].(types.Bool).ValueBool()
	multipleChoice := attrs["multiple_choice"].(types.Bool).ValueBool()

	if !closedPredefinedValues && multipleChoice {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Conflicting attribute values",
			"Setting closed_predefined_values to 'false' and multiple_choice to 'true' disables multiple_choice",
		)
	}
}
