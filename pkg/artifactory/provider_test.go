package artifactory

import (
	"context"
	"os"
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

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ARTIFACTORY_URL"); v == "" {
		t.Fatal("ARTIFACTORY_URL must be set for acceptance tests")
	}

	username := os.Getenv("ARTIFACTORY_USERNAME")
	password := os.Getenv("ARTIFACTORY_PASSWORD")
	api := os.Getenv("ARTIFACTORY_APIKEY")
	accessToken := os.Getenv("ARTIFACTORY_ACCESS_TOKEN")

	if (username == "" || password == "") && api == "" && accessToken == "" {
		t.Fatal("either ARTIFACTORY_USERNAME/ARTIFACTORY_PASSWORD or ARTIFACTORY_APIKEY  or ARTIFACTORY_ACCESS_TOKEN must be set for acceptance test")
	}

	ctx := context.Background()
	err := testAccProvider.Configure(ctx, terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}
