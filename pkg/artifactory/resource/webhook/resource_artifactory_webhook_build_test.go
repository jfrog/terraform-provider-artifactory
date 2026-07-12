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

package webhook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

var buildTemplate = `
	resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
		key         = "{{ .webhookName }}"
		description = "test description"
		event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
		criteria {
			any_build = false
			selected_builds = []
		}
		handler {
			url = "https://google.com"
		}
	}
`

func TestAccWebhook_Build_NilHandling(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_build_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhook_Build_NilHandling", `
		resource "artifactory_build_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["uploaded"]
			criteria {
				any_build  = true
				selected_builds = []
			}
			handler {
				url = "https://google.com"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "criteria.0.any_build", "true"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.selected_builds.#", "0"),
				),
			},
		},
	})
}

func TestAccWebhook_Build_UpgradeFromSDKv2(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_build_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	config := util.ExecuteTemplate("TestAccWebhook_Build_UpgradeFromSDKv2", `
		resource "artifactory_build_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["uploaded"]
			criteria {
				any_build  = false
				selected_builds = []
				include_patterns = ["foo"]
			}
			handler {
				url = "https://google.com"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		CheckDestroy: acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),

		Steps: []resource.TestStep{
			{
				Config: config,
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						Source:            "jfrog/artifactory",
						VersionConstraint: "12.1.0",
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo"),
				),
			},
			{
				Config:                   config,
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccWebhook_Build(t *testing.T) {
	id := testutil.RandomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_build_webhook.%s", name)

	params := map[string]interface{}{
		"webhookName": name,
	}
	webhookConfig := util.ExecuteTemplate("TestAccWebhookBuildPatterns", `
		resource "artifactory_build_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			event_types = ["uploaded"]
			criteria {
				any_build  = false
				selected_builds = []
				include_patterns = ["foo"]
			}
			handler {
				url = "https://google.com"
			}
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "criteria.0.include_patterns.0", "foo"),
				),
			},
		},
	})
}
