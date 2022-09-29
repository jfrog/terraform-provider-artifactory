package configuration_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccPropertySetCreate(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("property-set-", "artifactory_property_set")
	var testData = map[string]string{
		"resource_name":     resourceName,
		"property_set_name": "property-set-test",
		"visible":           "true",
		"property1":         "set1property1",
		"property2":         "set1property2",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccPropertySetDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, PropertySetTemplate, testData),
				Check:  resource.ComposeTestCheckFunc(verifyPropertySet(fqrn, testData)),
			},
		},
	})
}

func TestAccPropertySetUpdate(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("property-set-", "artifactory_property_set")
	var testData = map[string]string{
		"resource_name":     resourceName,
		"property_set_name": "property-set-test",
		"visible":           "false",
		"property1":         "set1property1-upd",
		"property2":         "set1property2-upd",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccPropertySetDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, PropertySetTemplate, testData),
				Check:  resource.ComposeTestCheckFunc(verifyPropertySet(fqrn, testData)),
			},
		},
	})
}

func TestAccPropertySetCustomizeDiff(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("property-set-", "artifactory_property_set")
	var testData = map[string]string{
		"resource_name":     resourceName,
		"property_set_name": "property-set-test",
		"visible":           "false",
		"property1":         "set1property1",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccPropertySetDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config:      util.ExecuteTemplate(fqrn, PropertySetCustomizeDiffTemplate, testData),
				ExpectError: regexp.MustCompile("setting closed_predefined_values to 'false' and multiple_choice to 'true' disables multiple_choice"),
			},
		},
	})
}

func verifyPropertySet(fqrn string, testData map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "name", testData["property_set_name"]),
		resource.TestCheckResourceAttr(fqrn, "visible", testData["visible"]),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.*", testData["property1"]),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.*", testData["property2"]),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.predefined_value.*.*", "passed-QA"),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.predefined_value.*.*", "failed-QA"),
	)
}

func testAccPropertySetDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources["artifactory_property_set."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		propertySets := &configuration.PropertySets{}

		response, err := client.R().SetResult(&propertySets).Get("artifactory/api/system/configuration")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		for _, iterPropertySet := range propertySets.PropertySets {
			if iterPropertySet.Name == id {
				return fmt.Errorf("error: Property set with key: " + id + " still exists.")
			}
		}
		return nil
	}
}

const PropertySetTemplate = `
resource "artifactory_property_set" "{{ .resource_name }}" {
  name 		= "{{ .property_set_name }}"
  visible 	= {{ .visible }}

  property {
      name = "{{ .property1 }}"

      predefined_value {
        name 			= "passed-QA"
        default_value 	= true
      }

      predefined_value {
        name 			= "failed-QA"
        default_value 	= false 
      }

      closed_predefined_values 	= true
      multiple_choice 			= true
  }

  property {
      name = "{{ .property2 }}"
    
      predefined_value {
        name 			= "passed-QA"
        default_value 	= true
      }

      predefined_value {
        name 			= "failed-QA"
        default_value 	= false 
      }

      closed_predefined_values 	= false
      multiple_choice 			= false
  }
}`

const PropertySetCustomizeDiffTemplate = `
resource "artifactory_property_set" "{{ .resource_name }}" {
  name 		= "{{ .property_set_name }}"
  visible 	= {{ .visible }}

  property {
      name = "{{ .property1 }}"

      predefined_value {
        name 			= "passed-QA"
        default_value 	= true
      }

      predefined_value {
        name 			= "failed-QA"
        default_value 	= false 
      }

      closed_predefined_values 	= false
      multiple_choice 			= true
  }
}`
