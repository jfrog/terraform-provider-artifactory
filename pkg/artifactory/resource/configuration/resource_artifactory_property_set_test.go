package configuration_test

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/configuration"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
)

func TestAccPropertySet(t *testing.T) {
	const PropertySet = `
resource "artifactory_property_set" "foo" {
  name 		= "property-set1"
  visible 	= true

  property {
      name = "set1property1"

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
      name = "set1property2"
    
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

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccPropertySetDestroy("foo"),

		Steps: []resource.TestStep{
			{
				Config: PropertySet,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_property_set.foo", "name", "property-set1"),
					resource.TestCheckResourceAttr("artifactory_property_set.foo", "visible", "true"),
					resource.TestCheckTypeSetElemAttr("artifactory_property_set.foo", "property.*.*", "set1property1"),
					resource.TestCheckTypeSetElemAttr("artifactory_property_set.foo", "property.*.*", "set1property2"),
					resource.TestCheckTypeSetElemAttr("artifactory_property_set.foo", "property.*.predefined_value.*.*", "passed-QA"),
					resource.TestCheckTypeSetElemAttr("artifactory_property_set.foo", "property.*.predefined_value.*.*", "failed-QA"),
				),
			},
		},
	})
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

		for _, iterLdapSetting := range propertySets.PropertySets {
			if iterLdapSetting.Name == id {
				return fmt.Errorf("error: Property set with key: " + id + " still exists.")
			}
		}
		return nil
	}
}
