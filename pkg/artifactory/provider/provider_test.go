package provider_test

import (
	// "io/ioutil"
	// "os"
	"testing"

	// "github.com/go-resty/resty/v2"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/provider"
	// "github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
)

func TestProvider(t *testing.T) {
	if err := provider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = provider.Provider()
}

//
// func testAccPreCheck(t *testing.T) {
// 	restyClient := acctest.GetTestResty(t)
//
// 	// Set customer base URL so repos that relies on it will work
// 	// https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-UpdateCustomURLBase
// 	_, err := restyClient.R().SetBody(os.Getenv("ARTIFACTORY_URL")).SetHeader("Content-Type", "text/plain").Put("/artifactory/api/system/configuration/baseUrl")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	provider, _ := acctest.TestAccProviders(Provider())["artifactory"]()
// 	_, oldErr := acctest.ConfigureProvider(provider)
// 	if oldErr != nil {
// 		t.Fatal(oldErr)
// 	}
// }
