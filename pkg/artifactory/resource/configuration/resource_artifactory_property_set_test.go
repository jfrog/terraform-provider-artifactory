package configuration_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func TestAccPropertySetCreate(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("property-set-", "artifactory_property_set")
	var testData = map[string]string{
		"resource_name":     resourceName,
		"property_set_name": resourceName,
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
				Config: utilsdk.ExecuteTemplate(fqrn, PropertySetTemplate, testData),
				Check:  resource.ComposeTestCheckFunc(verifyPropertySet(fqrn, testData)),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPropertySetUpdate(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("property-set-", "artifactory_property_set")
	var testData = map[string]string{
		"resource_name":            resourceName,
		"property_set_name":        resourceName,
		"visible":                  "false",
		"property1":                "set1property1",
		"default_value1":           "false",
		"default_value2":           "false",
		"closed_predefined_values": "true",
		"multiple_choice":          "true",
	}
	var testDataUpdated = map[string]string{
		"resource_name":            resourceName,
		"property_set_name":        resourceName,
		"visible":                  "false",
		"property1":                "set1property1",
		"default_value1":           "true",
		"default_value2":           "false",
		"closed_predefined_values": "true",
		"multiple_choice":          "false",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccPropertySetDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate(fqrn, PropertySetUpdateAndDiffTemplate, testData),
				Check:  resource.ComposeTestCheckFunc(verifyPropertySetUpdate(fqrn, testData)),
			},
			{
				Config: utilsdk.ExecuteTemplate(fqrn, PropertySetUpdateAndDiffTemplate, testDataUpdated),
				Check:  resource.ComposeTestCheckFunc(verifyPropertySetUpdate(fqrn, testDataUpdated)),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccPropertySetCustomizeDiff(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("property-set-", "artifactory_property_set")
	var testData = map[string]string{
		"resource_name":            resourceName,
		"property_set_name":        resourceName,
		"visible":                  "false",
		"property1":                "set1property1",
		"default_value1":           "false",
		"default_value2":           "false",
		"closed_predefined_values": "false",
		"multiple_choice":          "true",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccPropertySetDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config:      utilsdk.ExecuteTemplate(fqrn, PropertySetUpdateAndDiffTemplate, testData),
				ExpectError: regexp.MustCompile("setting closed_predefined_values to 'false' and multiple_choice to 'true' disables multiple_choice"),
			},
			{
				ResourceName:  fqrn,
				ImportStateId: resourceName,
				ImportState:   true,
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func TestAccPropertySet_importNotFound(t *testing.T) {
	config := `
		resource "artifactory_property_set" "not-exist-test" {
		  name                     = "not-exist-test"
		  visible                  = true
		  closed_predefined_values = true
		  multiple_choice          = true

		  property {
		    name = "property1"

		    predefined_value {
		      name          = "passed-QA"
		      default_value = true
	        }
		  }
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        config,
				ResourceName:  "artifactory_property_set.not-exist-test",
				ImportStateId: "not-exist-test",
				ImportState:   true,
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
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

func verifyPropertySetUpdate(fqrn string, testData map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "name", testData["property_set_name"]),
		resource.TestCheckResourceAttr(fqrn, "visible", testData["visible"]),
		resource.TestCheckResourceAttr(fqrn, "property.0.name", testData["property1"]),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.predefined_value.*.*", "passed-QA"),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.predefined_value.*.*", "failed-QA"),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.predefined_value.*.*", testData["default_value1"]),
		resource.TestCheckTypeSetElemAttr(fqrn, "property.*.predefined_value.*.*", testData["default_value2"]),
		resource.TestCheckResourceAttr(fqrn, "property.0.closed_predefined_values", testData["closed_predefined_values"]),
		resource.TestCheckResourceAttr(fqrn, "property.0.multiple_choice", testData["multiple_choice"]),
	)
}

func testAccPropertySetDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

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
      name          = "passed-QA"
      default_value = true
    }

    predefined_value {
      name          = "failed-QA"
      default_value = false
    }

    closed_predefined_values = true
    multiple_choice          = true
  }

  property {
    name = "{{ .property2 }}"

    predefined_value {
      name          = "passed-QA"
      default_value = true
    }

    predefined_value {
      name          = "failed-QA"
      default_value = false
    }

    closed_predefined_values = false
    multiple_choice          = false
  }
}`

const PropertySetUpdateAndDiffTemplate = `
resource "artifactory_property_set" "{{ .resource_name }}" {
  name    = "{{ .property_set_name }}"
  visible = {{ .visible }}

  property {
    name = "{{ .property1 }}"

    predefined_value {
      name          = "passed-QA"
      default_value = {{ .default_value1 }}
    }

    predefined_value {
      name          = "failed-QA"
      default_value = {{ .default_value2 }}
    }

    closed_predefined_values = {{ .closed_predefined_values }}
    multiple_choice          = {{ .multiple_choice }}
  }
}`
