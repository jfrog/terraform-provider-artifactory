package configuration

import (
	"context"
	"github.com/jfrog/terraform-provider-shared/packer"
	"github.com/jfrog/terraform-provider-shared/predicate"
	"gopkg.in/yaml.v3"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
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
	var resourcePropertySetRead = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		propertySetConfigs := &PropertySets{}
		// Unpacking HCL to compare the names of the property sets with the XML data we will get from the API
		unpackedPropertySet := unpackPropertySet(d)

		_, err := m.(*resty.Client).R().SetResult(&propertySetConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		var matchedPropertySet = PropertySet{}

		matchedPropertySet = PropertySet{}
		for _, iterPropertySet := range propertySetConfigs.PropertySets {
			if iterPropertySet.Name == unpackedPropertySet.Name {
				matchedPropertySet = iterPropertySet
				break
			}
		}

		pkr := packer.Universal(
			predicate.All(
				predicate.SchemaHasKey(propertySetsSchema),
			),
		)

		return diag.FromErr(pkr(&matchedPropertySet, d))
	}

	var parsePredefinedValues = func(values []PredefinedValue) map[string]interface{} {
		parsedPredefinedValues := map[string]interface{}{}
		for _, value := range values {
			parsedPredefinedValues[value.Name] = map[string]interface{}{
				"defaultValue": value.DefaultValue,
			}
		}
		return parsedPredefinedValues
	}

	var parseProperties = func(properties []Property) map[string]interface{} {
		parsedProperties := map[string]interface{}{}

		for _, property := range properties {
			parsedProperties[property.Name] = map[string]interface{}{
				"predefinedValues":       parsePredefinedValues(property.PredefinedValues),
				"closedPredefinedValues": property.ClosedPredefinedValue,
				"multipleChoice":         property.MultipleChoice,
			}
		}
		return parsedProperties
	}

	var resourcePropertySetsUpdate = func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		unpackedPropertySet := unpackPropertySet(d)

		///* EXPLANATION FOR BELOW CONSTRUCTION USAGE.
		//There is a difference in xml structure usage between GET and PATCH calls of API: /artifactory/api/system/configuration.
		//GET call structure has "propertySets -> propertySet -> Array of property sets".
		//PATCH call structure has "propertySets -> propertySet (dynamic sting). Property name and predefinedValues names are also dynamic strings".
		//Following nested map of string structs are constructed to match the usage of PATCH call with the consideration of dynamic strings.
		//*/
		var constructBody = map[string]map[string]map[string]interface{}{
			"propertySets": {
				unpackedPropertySet.Name: {
					"visible":    unpackedPropertySet.Visible,
					"properties": parseProperties(unpackedPropertySet.Properties),
				},
			},
		}

		content, err := yaml.Marshal(&constructBody)

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

	var resourceLdapSettingsDelete = func(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		propertySetConfigs := &PropertySets{}
		unpackedPropertySet := unpackPropertySet(d)

		response, err := m.(*resty.Client).R().SetResult(&propertySetConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return diag.Errorf("failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return diag.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}
		var matchedPropertySet = PropertySet{} // TODO: do we match before delete or just delete based on HCL list of sets?

		matchedPropertySet = PropertySet{}
		for _, iterPropertySet := range propertySetConfigs.PropertySets {
			if iterPropertySet.Name == unpackedPropertySet.Name {
				matchedPropertySet = iterPropertySet
				break
			}
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

		return nil
	}

	return &schema.Resource{
		UpdateContext: resourcePropertySetsUpdate,
		CreateContext: resourcePropertySetsUpdate,
		DeleteContext: resourceLdapSettingsDelete,
		ReadContext:   resourcePropertySetRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema:      propertySetsSchema,
		Description: "Provides an Artifactory Property Set resource. This resource configuration corresponds to 'propertySets' config block in system configuration XML (REST endpoint: artifactory/api/system/configuration).",
	}
}

func unpackPredefinedValues(s interface{}) []PredefinedValue {
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

func unpackPropertySet(s *schema.ResourceData) PropertySet {
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
