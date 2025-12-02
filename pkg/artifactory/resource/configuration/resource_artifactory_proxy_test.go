// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
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

func TestAccProxy_UpgradeFromSDKv2(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	providerHost := os.Getenv("TF_ACC_PROVIDER_HOST")
	if providerHost == "registry.opentofu.org" {
		t.Skipf("provider host is registry.opentofu.org. Previous version of Artifactory provider is unknown to OpenTofu.")
	}

	_, fqrn, resourceName := testutil.MkNames("test-proxy-", "artifactory_proxy")

	temp := `
	resource "artifactory_proxy" "{{ .resource_name }}" {
		key  = "{{ .resource_name }}"
		host = "{{ .host }}"
		port = {{ .port }}
	}`

	testData := map[string]string{
		"resource_name": resourceName,
		"host":          "https://fake-proxy.org",
		"port":          "8080",
	}

	config := util.ExecuteTemplate(fqrn, temp, testData)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "10.1.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", testData["resource_name"]),
					resource.TestCheckResourceAttr(fqrn, "host", testData["host"]),
					resource.TestCheckResourceAttr(fqrn, "port", testData["port"]),
					resource.TestCheckResourceAttr(fqrn, "username", ""),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckResourceAttr(fqrn, "nt_host", ""),
					resource.TestCheckResourceAttr(fqrn, "nt_domain", ""),
					resource.TestCheckResourceAttr(fqrn, "platform_default", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "redirect_to_hosts"),
					resource.TestCheckNoResourceAttr(fqrn, "services"),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccProxy_createUpdate(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, resourceName := testutil.MkNames("proxy-", "artifactory_proxy")
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccProxyDestroy(resourceName),

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
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        resourceName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "key",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

func TestAccProxy_importNotFound(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
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

func TestAccProxy_configValidation(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("proxy-", "artifactory_proxy")
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccProxyDestroy(resourceName),

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
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_proxy."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		proxies := &configuration.ProxiesAPIModel{}

		response, err := client.R().SetResult(&proxies).Get(configuration.ConfigurationEndpoint)
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		matchedProxyConfig := configuration.FindConfigurationById[configuration.ProxyAPIModel](proxies.Proxies, id)
		if matchedProxyConfig != nil {
			return fmt.Errorf("error: Proxy with key: %s still exists.", id)
		}

		return nil
	}
}
