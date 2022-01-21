package artifactory

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders = func() map[string]func() (*schema.Provider, error) {
	provider := Provider()
	return map[string]func() (*schema.Provider, error){
		"artifactory": func() (*schema.Provider, error) {
			return provider, nil
		},
	}
}()

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

func getTestResty(t *testing.T) *resty.Client {
	if v := os.Getenv("ARTIFACTORY_URL"); v == "" {
		t.Fatal("ARTIFACTORY_URL must be set for acceptance tests")
	}
	restyClient, err := buildResty(os.Getenv("ARTIFACTORY_URL"))
	if err != nil {
		t.Fatal(err)
	}
	username := os.Getenv("ARTIFACTORY_USERNAME")
	password := os.Getenv("ARTIFACTORY_PASSWORD")
	api := os.Getenv("ARTIFACTORY_APIKEY")
	accessToken := os.Getenv("ARTIFACTORY_ACCESS_TOKEN")
	restyClient, err = addAuthToResty(restyClient, username, password, api, accessToken)
	if err != nil {
		t.Fatal(err)
	}
	return restyClient
}

func testAccPreCheck(t *testing.T) {
	restyClient := getTestResty(t)

	// Set customer base URL so repos that relies on it will work
	// https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-UpdateCustomURLBase
	_, err := restyClient.R().SetBody(os.Getenv("ARTIFACTORY_URL")).SetHeader("Content-Type", "text/plain").Put("/artifactory/api/system/configuration/baseUrl")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	provider, _ := testAccProviders["artifactory"]()
	oldErr := provider.Configure(ctx, terraform.NewResourceConfigRaw(nil))
	if oldErr != nil {
		t.Fatal(oldErr)
	}
}
