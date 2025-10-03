package configuration_test

import (
    "os"
    "strings"
    "testing"
    "text/template"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
    "github.com/jfrog/terraform-provider-shared/testutil"
)

const baseUrlTemplate = `
resource "artifactory_configuration_base_url" "{{ .ResourceName }}" {
  base_url = "{{ .BaseURL }}"
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

    resource.Test(t, resource.TestCase{
        PreCheck: func() { acctest.PreCheck(t) },
        ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: formatBaseUrlConfig(resourceName, testData["base_url"]),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testCheckBaseUrlExists(fqrn, testData["base_url"]),
                ),
            },
            {
                Config: formatBaseUrlConfig(resourceName, "http://updated-test.com"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    testCheckBaseUrlExists(fqrn, "http://updated-test.com"),
                ),
            },
        },
    })
}

// formatBaseUrlConfig formats the Terraform configuration based on resource name and base URL
func formatBaseUrlConfig(resourceName, baseURL string) string {
    data := struct {
        ResourceName string
        BaseURL      string
    }{
        ResourceName: resourceName,
        BaseURL:      baseURL,
    }
    tmpl, err := template.New("baseUrl").Parse(baseUrlTemplate)
    if err != nil {
        panic(err) 
    }
    var result strings.Builder
    if err := tmpl.Execute(&result, data); err != nil {
        panic(err) 
    }

    return result.String()
}

// testCheckBaseUrlExists checks if the base URL exists and matches the expected value.
func testCheckBaseUrlExists(fqrn string, expectedBaseUrl string) resource.TestCheckFunc {
    return resource.ComposeTestCheckFunc(
        resource.TestCheckResourceAttr(fqrn, "base_url", expectedBaseUrl),
    )
}