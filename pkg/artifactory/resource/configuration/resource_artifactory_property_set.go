package configuration

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
	"gopkg.in/yaml.v3"
)

type PredefinedValue struct {
	Name         string `xml:"value" yaml:"-"`
	DefaultValue bool   `xml:"defaultValue" yaml:"defaultValue"`
}

type Property struct {
	Name                  string            `xml:"name" yaml:"-"`
	PredefinedValues      []PredefinedValue `xml:"predefinedValues>predefinedValue" yaml:"predefinedValues"`
	ClosedPredefinedValue bool              `xml:"closedPredefinedValues" yaml:"closedPredefinedValues"`
	MultipleChoice        bool              `xml:"multipleChoice" yaml:"multipleChoice"`
}

type PropertySet struct {
	Name       string     `xml:"name" yaml:"-"`
	Visible    bool       `xml:"visible" yaml:"visible"`
	Properties []Property `xml:"properties>property" yaml:"properties"`
}

func (p PropertySet) Id() string {
	return p.Name
}

type PropertySets struct {
	PropertySets []PropertySet `xml:"propertySets>propertySet" yaml:"propertySet"`
}

func ResourceArtifactoryPropertySet() *schema.Resource {
	var predefinedValueSchema = schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "Properties in the property set.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:             schema.TypeString,
					Required:         true,
					Description:      "Predefined property name.",
					ValidateDiagFunc: validator.StringIsNotEmpty,
				},
				"default_value": {
					Type:        schema.TypeBool,
					Required:    true,
					Description: "Whether the value is selected by default in the UI.",
				},
			},
		},
	}

	var propertySetsSchema = map[string]*schema.Schema{
		"name": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      "Property set name.",
		},
		"visible": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Defines if the list visible and assignable to the repository or artifact.",
		},
		"property": {
			Type:        schema.TypeSet,
			Required:    true,
			MinItems:    1,
			Description: "A list of properties that will be part of the property set.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "The name of the property.",
						ValidateDiagFunc: validator.StringIsNotEmpty,
					},
					"closed_predefined_values": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: `Disables "multiple_choice" if set to "false" at the same time with multiple_choice set to "true".`,
					},
					"multiple_choice": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: `Whether or not user can select multiple values. "closed_predefined_values" should be set to "true".`,
					},
					"predefined_value": &predefinedValueSchema,
				},
			},
		},
	}

	var unpackPredefinedValues = func(s interface{}) []PredefinedValue {
		predefinedValues := s.(*schema.Set).List()
		var values []PredefinedValue

		for _, v := range predefinedValues {

			id := v.(map[string]interface{})

			value := PredefinedValue{
				Name:         id["name"].(string),
				DefaultValue: id["default_value"].(bool),
			}
			values = append(values, value)
		}

		return values
	}

	var unpackPropertySet = func(s *schema.ResourceData) PropertySet {
		d := &util.ResourceData{ResourceData: s}
		propertySet := PropertySet{
			Name:    d.GetString("name", false),
			Visible: d.GetBool("visible", false),
		}

		var properties []Property

		if v, ok := d.GetOk("property"); ok {
			sets := v.(*schema.Set).List()
			if len(sets) == 0 {
				return propertySet
			}

			for _, set := range sets {
				id := set.(map[string]interface{})

				property := Property{
					Name:                  id["name"].(string),
					PredefinedValues:      unpackPredefinedValues(id["predefined_value"]),
					ClosedPredefinedValue: id["closed_predefined_values"].(bool),
					MultipleChoice:        id["multiple_choice"].(bool),
				}
				properties = append(properties, property)
			}
			propertySet.Properties = properties
		}

		return propertySet
	}

	var packPropertySet = func(p *PropertySet, d *schema.ResourceData) diag.Diagnostics {
		setValue := util.MkLens(d)

		setValue("name", p.Name)
		setValue("visible", p.Visible)

		var packPredefinedValues = func(predefinedValues []PredefinedValue) []interface{} {
			packedValues := []interface{}{}

			for _, predefinedValue := range predefinedValues {
				value := map[string]interface{}{
					"name":          predefinedValue.Name,
					"default_value": predefinedValue.DefaultValue,
				}

				packedValues = append(packedValues, value)
			}

			return packedValues
		}

		predefinedValueResource := predefinedValueSchema.Elem.(*schema.Resource)
		properties := []interface{}{}
		for _, prop := range p.Properties {
			property := map[string]interface{}{
				"name":                     prop.Name,
				"closed_predefined_values": prop.ClosedPredefinedValue,
				"multiple_choice":          prop.MultipleChoice,
				"predefined_value":         schema.NewSet(schema.HashResource(predefinedValueResource), packPredefinedValues(prop.PredefinedValues)),
			}

			properties = append(properties, property)
		}

		propertyResource := propertySetsSchema["property"].Elem.(*schema.Resource)
		errors := setValue("property", schema.NewSet(schema.HashResource(propertyResource), properties))

		if errors != nil && len(errors) > 0 {
			return diag.Errorf("failed to pack property_set %q", errors)
		}
		return nil
	}

	var resourcePropertySetRead = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		data := &util.ResourceData{ResourceData: d}
		name := data.GetString("name", false)

		propertySetConfigs := PropertySets{}

		_, err := m.(util.ProvderMetadata).Client.R().SetResult(&propertySetConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}

		matchedPropertySet := FindConfigurationById[PropertySet](propertySetConfigs.PropertySets, name)
		if matchedPropertySet == nil {
			d.SetId("")
			return nil
		}

		return packPropertySet(matchedPropertySet, d)
	}

	var transformPredefinedValues = func(values []PredefinedValue) map[string]interface{} {
		transformedPredefinedValues := map[string]interface{}{}
		for _, value := range values {
			transformedPredefinedValues[value.Name] = map[string]interface{}{
				"defaultValue": value.DefaultValue,
			}
		}
		return transformedPredefinedValues
	}

	var transformProperties = func(properties []Property) map[string]interface{} {
		transformedProperties := map[string]interface{}{}

		for _, property := range properties {
			transformedProperties[property.Name] = map[string]interface{}{
				"predefinedValues":       transformPredefinedValues(property.PredefinedValues),
				"closedPredefinedValues": property.ClosedPredefinedValue,
				"multipleChoice":         property.MultipleChoice,
			}
		}
		return transformedProperties
	}

	var resourcePropertySetsUpdate = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		unpackedPropertySet := unpackPropertySet(d)

		///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
		//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
		//GET call structure has "propertySets -> propertySet -> Array of property sets".
		//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
		//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
		//*/
		var body = map[string]map[string]map[string]interface{}{
			"propertySets": {
				unpackedPropertySet.Name: {
					"visible":    unpackedPropertySet.Visible,
					"properties": transformProperties(unpackedPropertySet.Properties),
				},
			},
		}

		content, err := yaml.Marshal(&body)

		if err != nil {
			return diag.Errorf("failed to marshal property set during Update")
		}

		err = SendConfigurationPatch(content, m)
		if err != nil {
			return diag.Errorf("failed to send PATCH request to Artifactory during Update")
		}

		d.SetId(unpackedPropertySet.Name)
		return resourcePropertySetRead(ctx, d, m)
	}

	var resourcePropertySetDelete = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		propertySetConfigs := &PropertySets{}

		response, err := m.(util.ProvderMetadata).Client.R().SetResult(&propertySetConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return diag.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		matchedPropertySet := FindConfigurationById[PropertySet](propertySetConfigs.PropertySets, d.Id())
		if matchedPropertySet == nil {
			return diag.Errorf("No property set found for '%s'", d.Id())
		}

		var constructBody = map[string]map[string]string{
			"propertySets": {
				matchedPropertySet.Name: "~",
			},
		}

		content, err := yaml.Marshal(&constructBody)
		if err != nil {
			return diag.Errorf("failed to marshal property set during Delete")
		}

		err = SendConfigurationPatch(content, m)
		if err != nil {
			return diag.Errorf("failed to send PATCH request to Artifactory during Delete")
		}

		d.SetId("")

		return nil
	}

	var verifyCrossDependentValues = func(ctx context.Context, diff *schema.ResourceDiff, v interface{}) error {
		if data, ok := diff.GetOk("property"); ok {
			sets := data.(*schema.Set).List()

			for _, set := range sets {
				id := set.(map[string]interface{})
				if id["closed_predefined_values"].(bool) == false && id["multiple_choice"].(bool) == true {
					return fmt.Errorf("setting closed_predefined_values to 'false' and multiple_choice to 'true' disables multiple_choice")
				}
			}
		}
		return nil
	}

	return &schema.Resource{
		UpdateContext: resourcePropertySetsUpdate,
		CreateContext: resourcePropertySetsUpdate,
		DeleteContext: resourcePropertySetDelete,
		ReadContext:   resourcePropertySetRead,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				d.Set("name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema:        propertySetsSchema,
		CustomizeDiff: verifyCrossDependentValues,
		Description:   "Provides an Artifactory Property Set resource. This resource configuration corresponds to 'propertySets' config block in system configuration XML (REST endpoint: artifactory/api/system/configuration).",
	}
}
