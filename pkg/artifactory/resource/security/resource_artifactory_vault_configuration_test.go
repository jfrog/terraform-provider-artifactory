package security_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/samber/lo"
)

// To execute this test suite, you need to:
// 1. Setup Vault dev server, create certificate, Vault policy, Vault role, etc. Follow
// steps here: https://jfrog.com/help/r/hashicorp-vault-setup-instructions/subject
// 2. Export env vars in your shell, e.g. VAULT_ADDR=http://host.docker.internal:8200
// 3. Ensure TLS is enabled (see access.config.patch.yml) and set env var JFROG_BYPASS_TLS_VERIFICATION=true
// 4. Start/restart your local Artifactory container (using /scripts/run-artifactory-container.sh)
// 5. Update JFROG_URL to use https

func TestAccVaultConfiguration_full(t *testing.T) {
	vaultAddr := os.Getenv("VAULT_ADDR")
	if len(vaultAddr) == 0 {
		t.Skipf("env var VAULT_ADDR is not set.")
	}

	vaultToken := os.Getenv("VAULT_TOKEN")
	if len(vaultToken) == 0 {
		t.Skipf("env var VAULT_TOKEN is not set.")
	}

	vaultRoleID := os.Getenv("VAULT_ROLE_ID")
	if len(vaultRoleID) == 0 {
		t.Skipf("env var VAULT_ROLE_ID is not set.")
	}

	vaultSecretID := os.Getenv("VAULT_SECRET_ID")
	if len(vaultSecretID) == 0 {
		t.Skipf("env var VAULT_SECRET_ID is not set.")
	}

	vaultPath := os.Getenv("VAULT_PATH")
	if len(vaultPath) == 0 {
		t.Skipf("env var VAULT_PATH is not set.")
	}

	_, fqrn, resourceName := testutil.MkNames("vault-config-", "artifactory_vault_configuration")

	const template = `
		resource "artifactory_vault_configuration" "{{ .name }}" {
			name = "{{ .name }}"
			config = {
				url = "{{ .url }}"
				auth = {
					type      = "AppRole"
					role_id   = "{{ .role_id }}"
					secret_id = "{{ .secret_id }}"
				}

				mounts = [
					{
						path = "{{ .path }}"
						type = "{{ .type }}"
					}
				]
			}
		}
	`

	testData := map[string]string{
		"name":      resourceName,
		"url":       vaultAddr,
		"role_id":   vaultRoleID,
		"secret_id": vaultSecretID,
		"type":      "KV2",
		"path":      vaultPath,
	}

	config := util.ExecuteTemplate("TestAccVaultConfiguration_full", template, testData)

	const updatedTemplate = `
		resource "artifactory_vault_configuration" "{{ .name }}" {
			name = "{{ .name }}"
			config = {
				url = "{{ .url }}"
				auth = {
					type = "Agent"
				}

				mounts = [
					{
						path = "{{ .path }}"
						type = "{{ .type }}"
					}
				]
			}
		}
	`
	updatedTestData := map[string]string{
		"name":      resourceName,
		"url":       vaultAddr,
		"role_id":   vaultRoleID,
		"secret_id": vaultSecretID,
		"type":      "KV2",
		"path":      vaultPath,
	}

	updatedConfig := util.ExecuteTemplate("TestAccVaultConfiguration_full", updatedTemplate, updatedTestData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy:             testAccVaultConfigurationDestroy(testData["name"]),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "config.url", testData["url"]),
					resource.TestCheckResourceAttr(fqrn, "config.auth.type", "AppRole"),
					resource.TestCheckResourceAttr(fqrn, "config.auth.role_id", testData["role_id"]),
					resource.TestCheckResourceAttr(fqrn, "config.auth.secret_id", testData["secret_id"]),
					resource.TestCheckResourceAttr(fqrn, "config.mounts.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "config.mounts.0.path", testData["path"]),
					resource.TestCheckResourceAttr(fqrn, "config.mounts.0.type", testData["type"]),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "config.url", testData["url"]),
					resource.TestCheckResourceAttr(fqrn, "config.auth.type", "Agent"),
					resource.TestCheckNoResourceAttr(fqrn, "config.auth.role_id"),
					resource.TestCheckNoResourceAttr(fqrn, "config.auth.secret_id"),
					resource.TestCheckResourceAttr(fqrn, "config.mounts.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "config.mounts.0.path", testData["path"]),
					resource.TestCheckResourceAttr(fqrn, "config.mounts.0.type", testData["type"]),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        testData["name"],
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func TestAccVaultConfiguration_missing_auth_attrs(t *testing.T) {
	testCase := []struct {
		authType string
		errMsg   string
	}{
		{"AppRole", `.*Expected 'role_id' to be configured when auth type set to 'AppRole'.*`},
		{"Certificate", `.*Expected 'certificate' to be configured when auth type set to 'Certificate'.*`},
	}

	for _, tc := range testCase {
		t.Run(tc.authType, testInvalidAuthAttrs(tc.authType, tc.errMsg))
	}
}

func testInvalidAuthAttrs(authType, errMsg string) func(t *testing.T) {
	return func(t *testing.T) {
		_, _, resourceName := testutil.MkNames("vault-config-", "artifactory_vault_configuration")

		const template = `
		resource "artifactory_vault_configuration" "{{ .name }}" {
			name = "{{ .name }}"
			config = {
				url = "https://tempurl.org"
				auth = {
					type = "{{ .type }}"
				}

				mounts = [
					{
						path = "test-path"
						type = "KV2"
					}
				]
			}
		}`

		testData := map[string]string{
			"name": resourceName,
			"type": authType,
		}

		config := util.ExecuteTemplate("TestAccVaultConfiguration_full", template, testData)

		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acctest.PreCheck(t) },
			ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
			CheckDestroy:             testAccVaultConfigurationDestroy(testData["name"]),
			Steps: []resource.TestStep{
				{
					Config:      config,
					ExpectError: regexp.MustCompile(errMsg),
				},
			},
		})
	}
}

func testAccVaultConfigurationDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_vault_configuration."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}

		var vaultConfigs []security.VaultConfigurationAPIModel

		response, err := client.R().
			SetResult(&vaultConfigs).
			Get(security.VaultConfigurationsEndpoint)
		if err != nil {
			return err
		}
		if response.IsError() {
			return fmt.Errorf(response.String())
		}

		_, ok = lo.Find(
			vaultConfigs,
			func(config security.VaultConfigurationAPIModel) bool {
				return config.Key == id
			},
		)
		if ok {
			return fmt.Errorf("error: Vault configuration %s still exists", id)
		}

		return nil
	}
}
