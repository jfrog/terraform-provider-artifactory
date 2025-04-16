package configuration_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
)

const baseUrlTemplate = `
resource "artifactory_configuration_base_url" "{{ .resource_name }}" {
  base_url = "{{ .base_url }}"
}`

// TestAccBaseUrl_CreateUpdate tests creating and updating the Base URL resource.
func TestAccBaseUrl_CreateUpdate(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}
	_, fqrn, resourceName := testutil.MkNames("baseUrl", "artifactory_configuration_base_url")

	testData := map[string]string{
		"base_url": "http://test.com",
	}

	// first as inital step, we se the http://test.com as the base URl, we update the baseURl and confrim base url were updated.
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acctest.PreCheck(t) },
		Providers: testutil.SupportedProviders,
		Steps: []resource.TestStep{
			// set http://test.com as baseurl
			{
				Config: formatBaseUrlConfig(resourceName, testData["base_url"]),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckBaseUrlExists(fqrn, testData["base_url"]),
				),
			},
			{
				Config: formatBaseUrlConfig(testData["base_url"], "http://updated-test.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckBaseUrlExists(fqrn, "http://updated-test.com"),
				),
			},
		},
	})
}

// formatBaseUrlConfig formats the Terraform configuration based on resource name and base URL
func formatBaseUrlConfig(resourceName, baseURL string) string {
	return fmt.Sprintf(baseUrlTemplate, map[string]string{
		"base_url": baseURL,
	})
}
