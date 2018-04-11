package artifactory

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"artifactory": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
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
	token := os.Getenv("ARTIFACTORY_TOKEN")
	password := os.Getenv("ARTIFACTORY_PASSWORD")

	if (username == "" || password == "") && token == "" {
		t.Fatal("either ARTIFACTORY_USERNAME/ARTIFACTORY_PASSWORD or ARTIFACTORY_TOKEN must be set for acceptance test")
	}

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}
