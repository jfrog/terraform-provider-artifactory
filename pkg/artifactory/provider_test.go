package artifactory

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

const rtDefaultUser = "admin"

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}

func uploadTestFile(client *resty.Client, localPath, remotePath, contentType string) error {
	body, err := ioutil.ReadFile(localPath)
	if err != nil {
		return err
	}
	uri := "/artifactory/" + remotePath
	_, err = client.R().SetBody(body).SetHeader("Content-Type", contentType).Put(uri)
	return err
}

func testAccPreCheck(t *testing.T) {
	restyClient := utils.GetTestResty(t)

	// Set customer base URL so repos that relies on it will work
	// https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-UpdateCustomURLBase
	_, err := restyClient.R().SetBody(os.Getenv("ARTIFACTORY_URL")).SetHeader("Content-Type", "text/plain").Put("/artifactory/api/system/configuration/baseUrl")
	if err != nil {
		t.Fatal(err)
	}

	provider, _ := utils.TestAccProviders(Provider())["artifactory"]()
	_, oldErr := utils.ConfigureProvider(provider)
	if oldErr != nil {
		t.Fatal(oldErr)
	}
}
