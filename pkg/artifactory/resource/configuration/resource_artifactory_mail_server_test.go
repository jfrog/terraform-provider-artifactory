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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccMailServer_full(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	_, fqrn, resourceName := testutil.MkNames("mailserver-", "artifactory_mail_server")

	const mailServerTemplate = `
	resource "artifactory_mail_server" "{{ .resourceName }}" {
		enabled         = true
		artifactory_url = "{{ .artifactory_url }}"
		from            = "{{ .from }}"
		host            = "{{ .host }}"
		username        = "test-user"
		password        = "test-password"
		port            = 25
		subject_prefix  = "[Test]"
	}`

	testData := map[string]string{
		"resourceName":    resourceName,
		"artifactory_url": "http://tempurl.org",
		"from":            "test@jfrog.com",
		"host":            "http://tempurl.org",
	}

	const mailServerTemplateUpdate = `
	resource "artifactory_mail_server" "{{ .resourceName }}" {
		enabled         = true
		artifactory_url = "{{ .artifactory_url }}"
		from            = "{{ .from }}"
		host            = "{{ .host }}"
		port            = 25
		subject_prefix  = "[Test]"
		use_ssl         = true
		use_tls         = true
	}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccMailServerDestroy(resourceName),

		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, mailServerTemplate, testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "artifactory_url", testData["artifactory_url"]),
					resource.TestCheckResourceAttr(fqrn, "from", testData["from"]),
					resource.TestCheckResourceAttr(fqrn, "host", testData["host"]),
					resource.TestCheckResourceAttr(fqrn, "username", "test-user"),
					resource.TestCheckResourceAttr(fqrn, "password", "test-password"),
					resource.TestCheckResourceAttr(fqrn, "port", "25"),
					resource.TestCheckResourceAttr(fqrn, "subject_prefix", "[Test]"),
					resource.TestCheckResourceAttr(fqrn, "use_ssl", "false"),
					resource.TestCheckResourceAttr(fqrn, "use_tls", "false"),
				),
			},
			{
				Config: util.ExecuteTemplate(fqrn, mailServerTemplateUpdate, testData),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "artifactory_url", testData["artifactory_url"]),
					resource.TestCheckResourceAttr(fqrn, "from", testData["from"]),
					resource.TestCheckResourceAttr(fqrn, "host", testData["host"]),
					resource.TestCheckNoResourceAttr(fqrn, "username"),
					resource.TestCheckNoResourceAttr(fqrn, "password"),
					resource.TestCheckResourceAttr(fqrn, "port", "25"),
					resource.TestCheckResourceAttr(fqrn, "subject_prefix", "[Test]"),
					resource.TestCheckResourceAttr(fqrn, "use_ssl", "true"),
					resource.TestCheckResourceAttr(fqrn, "use_tls", "true"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportStateId:                        resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "host",
				ImportStateVerifyIgnore:              []string{"password"},
			},
		},
	})
}

func TestAccMailServer_invalid_from(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("mailserver-", "artifactory_mail_server")

	template := `
	resource "artifactory_mail_server" "{{ .resourceName }}" {
		enabled         = true
		artifactory_url = "http://tempurl.org"
		from            = "invalid-email"
		host            = "http://tempurl.org"
		username        = "test-user"
		password        = "test-password"
		port            = 25
		subject_prefix  = "[Test]"
		use_ssl         = true
		use_tls         = true
	}`

	testData := map[string]string{
		"resourceName": resourceName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:       util.ExecuteTemplate(fqrn, template, testData),
				ResourceName: resourceName,
				ExpectError:  regexp.MustCompile("value must be a valid email address"),
			},
		},
	})
}

func TestAccMailServer_invalid_artifactory_url(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("mailserver-", "artifactory_mail_server")

	template := `
	resource "artifactory_mail_server" "{{ .resourceName }}" {
		enabled         = true
		artifactory_url = "invalid-url"
		from            = "test-user@jfrog.com"
		host            = "http://tempurl.org"
		username        = "test-user"
		password        = "test-password"
		port            = 25
		subject_prefix  = "[Test]"
		use_ssl         = true
		use_tls         = true
	}`

	testData := map[string]string{
		"resourceName": resourceName,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:       util.ExecuteTemplate(fqrn, template, testData),
				ResourceName: resourceName,
				ExpectError:  regexp.MustCompile("value must be a valid URL with host.*"),
			},
		},
	})
}

func testAccMailServerDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_mail_server."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		var mailServer configuration.MailServer

		response, err := client.R().SetResult(&mailServer).Get(configuration.ConfigurationEndpoint)
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read. Response:%#v", response)
		}

		if mailServer.Server != nil {
			return fmt.Errorf("error: MailServer config still exists.")
		}

		return nil
	}
}
