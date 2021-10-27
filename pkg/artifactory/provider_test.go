package artifactory

import (
	"context"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"os"
	"testing"

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
	// https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-UpdateCustomURLBase
	restyClient := getTestResty(t)
	// TODO check the payload and make sure it's the right license type
	_, err := restyClient.R().Get("/artifactory/api/system/licenses/")
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
