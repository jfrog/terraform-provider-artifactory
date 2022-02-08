package artifactory

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var domainRepoTypeLookup = map[string]string {
	"artifact": "generic",
	"artifact_property": "generic",
	"docker": "docker_v2",
}

var domainEventTypesLookup = map[string][]string {
	"artifact": []string{"copied", "deleted", "deployed", "moved"},
	"artifact_property": []string{"added", "deleted"},
	"docker": []string{"pushed", "deleted", "promoted"},
}

func TestAccWebhookAllTypes(t *testing.T) {
	// Can only realistically test these 3 types of webhook since creating
	// build, release_bundle, or distribution in test environment is almost impossible
	for _, webhookType := range []string{"artifact", "artifact_property", "docker"} {
		t.Run(fmt.Sprintf("TestWebhook%s", strings.Title(strings.ToLower(webhookType))), func(t *testing.T) {
			resource.Test(webhookTestCase(webhookType, t))
		})
	}
}

func webhookTestCase(webhookType string, t *testing.T) (*testing.T, resource.TestCase) {
	id := randomInt()
	name := fmt.Sprintf("webhook-%d", id)
	fqrn := fmt.Sprintf("artifactory_%s_webhook.%s", webhookType, name)

	repoType := domainRepoTypeLookup[webhookType]
	repoName := fmt.Sprintf("%s-local-%d", webhookType, id)
	eventTypes := domainEventTypesLookup[webhookType]

	params := map[string]interface{}{
		"repoName":    repoName,
		"repoType":    repoType,
		"webhookType": webhookType,
		"webhookName": name,
		"eventTypes":  eventTypes,
		"anyLocal":    randBool(),
		"anyRemote":   randBool(),
	}
	webhookConfig := executeTemplate("TestAccWebhook{{ .webhookType }}Type", `
		resource "artifactory_local_{{ .repoType }}_repository" "{{ .repoName }}" {
			key = "{{ .repoName }}"
		}

		resource "artifactory_{{ .webhookType }}_webhook" "{{ .webhookName }}" {
			key         = "{{ .webhookName }}"
			description = "test description"
			domain      = "{{ .webhookType }}"
			event_types = [{{ range $index, $eventType := .eventTypes}}{{if $index}},{{end}}"{{$eventType}}"{{end}}]
			criteria {
				any_local  = {{ .anyLocal }}
				any_remote = {{ .anyRemote }}
				repo_keys  = ["{{ .repoName }}"]
			}
			url    = "http://tempurl.org"
			secret = "fake-secret"
			custom_http_header {
				name  = "header-1"
				value = "value-1"
			}
			custom_http_header {
				name  = "header-2"
				value = "value-2"
			}

			depends_on = [artifactory_local_{{ .repoType }}_repository.{{ .repoName }}]
		}
	`, params)

	testChecks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(fqrn, "key", name),
		resource.TestCheckResourceAttr(fqrn, "domain", webhookType),
		resource.TestCheckResourceAttr(fqrn, "event_types.#", fmt.Sprintf("%d", len(eventTypes))),
		resource.TestCheckResourceAttr(fqrn, "url", "http://tempurl.org"),
		resource.TestCheckResourceAttr(fqrn, "secret", "fake-secret"),
		resource.TestCheckResourceAttr(fqrn, "criteria.#", "1"),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_local", fmt.Sprintf("%t", params["anyLocal"])),
		resource.TestCheckResourceAttr(fqrn, "criteria.0.any_remote", fmt.Sprintf("%t", params["anyRemote"])),
		resource.TestCheckTypeSetElemAttr(fqrn, "criteria.0.repo_keys.*", repoName),
		resource.TestCheckResourceAttr(fqrn, "custom_http_header.#", "2"),
		resource.TestCheckResourceAttr(fqrn, "custom_http_header.0.name", "header-1"),
		resource.TestCheckResourceAttr(fqrn, "custom_http_header.0.value", "value-1"),
		resource.TestCheckResourceAttr(fqrn, "custom_http_header.1.name", "header-2"),
		resource.TestCheckResourceAttr(fqrn, "custom_http_header.1.value", "value-2"),
	}

	for _, eventType := range eventTypes {
		eventTypeCheck := resource.TestCheckTypeSetElemAttr(fqrn, "event_types.*", eventType)
		testChecks = append(testChecks, eventTypeCheck)
	}

	return t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		CheckDestroy: verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: webhookConfig,
				Check: resource.ComposeTestCheckFunc(testChecks...),
			},
		},
	}
}
