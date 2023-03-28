package configuration_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
)

const ProxyTemplate = `
resource "artifactory_proxy" "{{ .resource_name }}" {
  key               = "{{ .resource_name }}"
  host              = "{{ .host }}"
  port              = {{ .port }}
  username          = "{{ .username }}"
  password          = "{{ .password }}"
  nt_host           = "{{ .nt_host }}"
  nt_domain         = "{{ .nt_domain }}"
  platform_default  = {{ .platform_default }}
  redirect_to_hosts = ["{{ .redirect_to_hosts_1 }}"]
}`

const ProxyUpdatedTemplate = `
resource "artifactory_proxy" "{{ .resource_name }}" {
  key               = "{{ .resource_name }}"
  host              = "{{ .host }}"
  port              = {{ .port }}
  username          = "{{ .username }}"
  password          = "{{ .password }}"
  nt_host           = "{{ .nt_host }}"
  nt_domain         = "{{ .nt_domain }}"
  platform_default  = {{ .platform_default }}
  redirect_to_hosts = ["{{ .redirect_to_hosts_1 }}", "{{ .redirect_to_hosts_2 }}"]
  services          = ["{{ .services_1 }}", "{{ .services_2 }}"]
}`

func TestAccProxyCreateUpdate(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("proxy-", "artifactory_proxy")
	var testData = map[string]string{
		"resource_name":       resourceName,
		"host":                "https://fake-proxy.org",
		"port":                "8080",
		"username":            "fake-user",
		"password":            "fake-password",
		"nt_host":             "test-nt-host",
		"nt_domain":           "test-nt-domain",
		"platform_default":    "true",
		"redirect_to_hosts_1": "foo",
	}
	var testDataUpdated = map[string]string{
		"resource_name":       resourceName,
		"host":                "https://fake-proxy.org",
		"port":                "8080",
		"username":            "fake-user",
		"password":            "fake-password",
		"nt_host":             "test-nt-host",
		"nt_domain":           "test-nt-domain",
		"platform_default":    "false",
		"redirect_to_hosts_1": "foo",
		"redirect_to_hosts_2": "bar",
		"services_1":          "jfrt",
		"services_2":          "jfxr",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccProxyDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, ProxyTemplate, testData),
				Check:  resource.ComposeTestCheckFunc(verifyProxy(fqrn, testData)),
			},
			{
				Config: util.ExecuteTemplate(fqrn, ProxyUpdatedTemplate, testDataUpdated),
				Check:  resource.ComposeTestCheckFunc(verifyProxy(fqrn, testDataUpdated)),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func TestAccProxy_importNotFound(t *testing.T) {
	config := `
		resource "artifactory_proxy" "not-exist-test" {
		  key               = "not-exist-test"
		  host              = "https://fake-proxy.org"
		  port              = 8080
		  username          = "fake-user"
		  password          = "fake-password"
		  nt_host           = "fake-nt-host"
		  nt_domain         = "fake-nt-domain"
		  platform_default  = false
		  redirect_to_hosts = ["foo"]
		  services          = ["jfrt"]
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        config,
				ResourceName:  "artifactory_proxy.not-exist-test",
				ImportStateId: "not-exist-test",
				ImportState:   true,
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func TestAccProxyCustomizeDiff(t *testing.T) {
	_, fqrn, resourceName := test.MkNames("proxy-", "artifactory_proxy")
	var testData = map[string]string{
		"resource_name":       resourceName,
		"host":                "https://fake-proxy.org",
		"port":                "8080",
		"username":            "fake-user",
		"password":            "fake-password",
		"nt_host":             "test-nt-host",
		"nt_domain":           "test-nt-domain",
		"platform_default":    "true",
		"redirect_to_hosts_1": "foo",
		"redirect_to_hosts_2": "bar",
		"services_1":          "jfrt",
		"services_2":          "jfxr",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccProxyDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config:      util.ExecuteTemplate(fqrn, ProxyUpdatedTemplate, testData),
				ExpectError: regexp.MustCompile("services cannot be set when platform_default is true"),
			},
		},
	})
}

func verifyProxy(fqrn string, testData map[string]string) resource.TestCheckFunc {
	checkFunc := resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(fqrn, "key", testData["resource_name"]),
		resource.TestCheckResourceAttr(fqrn, "host", testData["host"]),
		resource.TestCheckResourceAttr(fqrn, "port", testData["port"]),
		resource.TestCheckResourceAttr(fqrn, "username", testData["username"]),
		resource.TestCheckResourceAttr(fqrn, "password", testData["password"]),
		resource.TestCheckResourceAttr(fqrn, "nt_host", testData["nt_host"]),
		resource.TestCheckResourceAttr(fqrn, "nt_domain", testData["nt_domain"]),
		resource.TestCheckResourceAttr(fqrn, "platform_default", testData["platform_default"]),
	)

	if v, ok := testData["redirect_to_hosts_1"]; ok {
		checkFunc = resource.ComposeTestCheckFunc(
			checkFunc,
			resource.TestCheckTypeSetElemAttr(fqrn, "redirect_to_hosts.*", v),
		)
	}

	if v, ok := testData["redirect_to_hosts_2"]; ok {
		checkFunc = resource.ComposeTestCheckFunc(
			checkFunc,
			resource.TestCheckTypeSetElemAttr(fqrn, "redirect_to_hosts.*", v),
		)
	}

	if v, ok := testData["services_1"]; ok {
		checkFunc = resource.ComposeTestCheckFunc(
			checkFunc,
			resource.TestCheckTypeSetElemAttr(fqrn, "services.*", v),
		)
	}

	if v, ok := testData["services_2"]; ok {
		checkFunc = resource.ComposeTestCheckFunc(
			checkFunc,
			resource.TestCheckTypeSetElemAttr(fqrn, "services.*", v),
		)
	}

	return checkFunc
}

func testAccProxyDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProvderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_proxy."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		proxies := &configuration.Proxies{}

		response, err := client.R().SetResult(&proxies).Get("artifactory/api/system/configuration")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		for _, proxy := range proxies.Proxies {
			if proxy.Key == id {
				return fmt.Errorf("error: Proxy with key: " + id + " still exists.")
			}
		}
		return nil
	}
}
