package artifactory

import (
	"context"
	"github.com/go-resty/resty/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"artifactory": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ = Provider()
}
func uploadTestFile(client *resty.Client, localPath, remotePath, contentType string) error {
	uri := "/artifactory/" + remotePath
	_, err := client.R().SetFile(filepath.Base(localPath),localPath).
		SetHeader("Content-Type", contentType).Put(uri)
	//curl -n --location --request PUT 'http://localhost:8081/artifactory/example-repo-local/artifact.zip' \
	//> --header 'Content-Type: application/zip' \
	//> --data-binary '@/Users/christianb/go/pkg/mod/github.com/klauspost/compress@v1.11.2/zstd/testdata/good.zip'
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
	// TODO check the payload and make sure it's the right license type
	_, err := restyClient.R().Get("/artifactory/api/system/licenses/")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	oldErr := testAccProvider.Configure(ctx, terraform.NewResourceConfigRaw(nil))
	if oldErr != nil {
		t.Fatal(oldErr)
	}
}
