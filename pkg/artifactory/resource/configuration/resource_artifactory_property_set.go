package configuration

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
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
	PredefinedValues      []PredefinedValue `xml:"predefinedValues" yaml:"predefinedValues"`
	ClosedPredefinedValue bool              `xml:"closedPredefinedValues" yaml:"closedPredefinedValues"`
	MultipleChoice        bool              `xml:"multipleChoice" yaml:"multipleChoice"`
}

type PropertySet struct {
	Name       string     `xml:"name" yaml:"-"`
	Visible    bool       `xml:"visible" yaml:"visible"`
	Properties []Property `xml:"properties>property" yaml:"properties"`
}

type PropertySets struct {
	PropertySets []PropertySet `xml:"propertySets>propertySet" yaml:"propertySet"`
}

func ResourceArtifactoryPropertySet() *schema.Resource {
	var propertySetsSchema = map[string]*schema.Schema{
		"name": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validator.StringIsNotEmpty,
			Description:      `Property set name.`,
		},
		"visible": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: `Defines if the list visible and assignable to the repository or artifact.`,
		},
		"property": {
			Type:        schema.TypeSet,
			Required:    true,
			MinItems:    1,
			Description: `A list of properties that will be part of the property set.`,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      `The name of the property.`,
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
					"predefined_value": {
						Type:        schema.TypeSet,
						Required:    true,
						Description: `Properties in the property set.`,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type:             schema.TypeString,
									Required:         true,
									Description:      `Predefined property name.`,
									ValidateDiagFunc: validator.StringIsNotEmpty,
								},
								"default_value": {
									Type:        schema.TypeBool,
									Required:    true,
									Description: `Whether the value is selected by default in the UI.`,
								},
							},
						},
					},
				},
			},
		},
	}

	var findPropertySet = func(propertySets *PropertySets, name string) PropertySet {
		for _, propertySet := range propertySets.PropertySets {
			if propertySet.Name == name {
				return propertySet
			}
		}
		return PropertySet{}
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

	var resourcePropertySetRead = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		propertySetConfigs := &PropertySets{}
		// Unpacking HCL to compare the names of the property sets with the XML data we will get from the API
		unpackedPropertySet := unpackPropertySet(d)

		_, err := m.(*resty.Client).R().SetResult(&propertySetConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}

		matchedPropertySet := findPropertySet(propertySetConfigs, unpackedPropertySet.Name)

		pkr := packer.Universal(
			predicate.All(
				predicate.SchemaHasKey(propertySetsSchema),
			),
		)

		return diag.FromErr(pkr(&matchedPropertySet, d))
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
		unpackedPropertySet := unpackPropertySet(d)

		response, err := m.(*resty.Client).R().SetResult(&propertySetConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return diag.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		matchedPropertySet := findPropertySet(propertySetConfigs, unpackedPropertySet.Name)

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
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:        propertySetsSchema,
		CustomizeDiff: verifyCrossDependentValues,
		Description:   "Provides an Artifactory Property Set resource. This resource configuration corresponds to 'propertySets' config block in system configuration XML (REST endpoint: artifactory/api/system/configuration).",
	}
}
