package configuration_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccLdapGroupSettingV2_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("ldap-", "artifactory_ldap_group_setting_v2")

	const ldapGroupSetting = `
	resource "artifactory_ldap_group_setting_v2" "{{ .name }}" {
		name = "{{ .name }}"
		enabled_ldap = "{{ .enabled_ldap }}"
		group_base_dn = "{{ .group_base_dn }}"
		group_name_attribute = "cn"
		group_member_attribute = "{{ .group_member_attribute }}"
		sub_tree = true
		force_attribute_search = false
		filter = "(objectClass=groupOfNames)"
		description_attribute = "description"
		strategy = "{{ .strategy }}"
	}
	`
	params := map[string]interface{}{
		"name":                   name,
		"enabled_ldap":           "ldap2",
		"group_base_dn":          "CN=Users,DC=MyDomain,DC=com",
		"group_member_attribute": "uniqueMember",
		"strategy":               "STATIC",
	}
	LdapSettingTemplateFull := utilsdk.ExecuteTemplate("TestLdap", ldapGroupSetting, params)

	paramsUpdate := map[string]interface{}{
		"name":                   name,
		"enabled_ldap":           "ldap3",
		"group_base_dn":          "CN=Users,DC=MyDomain,DC=org",
		"group_member_attribute": "uniqueMember1",
		"strategy":               "DYNAMIC",
	}
	LdapSettingTemplateFullUpdate := utilsdk.ExecuteTemplate("TestLdap", ldapGroupSetting, paramsUpdate)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccLdapGroupSettingV2Destroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: LdapSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "enabled_ldap", "ldap2"),
					resource.TestCheckResourceAttr(fqrn, "group_base_dn", "CN=Users,DC=MyDomain,DC=com"),
					resource.TestCheckResourceAttr(fqrn, "group_name_attribute", "cn"),
					resource.TestCheckResourceAttr(fqrn, "group_member_attribute", "uniqueMember"),
					resource.TestCheckResourceAttr(fqrn, "sub_tree", "true"),
					resource.TestCheckResourceAttr(fqrn, "force_attribute_search", "false"),
					resource.TestCheckResourceAttr(fqrn, "filter", "(objectClass=groupOfNames)"),
					resource.TestCheckResourceAttr(fqrn, "description_attribute", "description"),
					resource.TestCheckResourceAttr(fqrn, "strategy", "STATIC"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				Config: LdapSettingTemplateFullUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "enabled_ldap", "ldap3"),
					resource.TestCheckResourceAttr(fqrn, "group_base_dn", "CN=Users,DC=MyDomain,DC=org"),
					resource.TestCheckResourceAttr(fqrn, "group_name_attribute", "cn"),
					resource.TestCheckResourceAttr(fqrn, "group_member_attribute", "uniqueMember1"),
					resource.TestCheckResourceAttr(fqrn, "sub_tree", "true"),
					resource.TestCheckResourceAttr(fqrn, "force_attribute_search", "false"),
					resource.TestCheckResourceAttr(fqrn, "filter", "(objectClass=groupOfNames)"),
					resource.TestCheckResourceAttr(fqrn, "description_attribute", "description"),
					resource.TestCheckResourceAttr(fqrn, "strategy", "DYNAMIC"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "name"),
			},
		},
	})
}

func TestAccLdapGroupSettingV2_failingValidators(t *testing.T) {
	_, _, name := testutil.MkNames("ldap-", "artifactory_ldap_group_setting_v2")
	errorMessageConfiguration := "Incorrect Attribute Configuration"
	paramsConflict := map[string]interface{}{
		"name":                   name,
		"enabled_ldap":           "ldap2",
		"group_base_dn":          "CN=Users,DC=MyDomain,DC=com",
		"group_member_attribute": "uniqueMember",
		"sub_tree":               "true",
		"strategy":               "HIERARCHICAL",
	}
	t.Run(fmt.Sprintf("TestLdapGroup_ConflictStrategySubTree"), func(t *testing.T) {
		resource.Test(makeLdapGroupValidatorsTestCase(paramsConflict, errorMessageConfiguration, t))
	})

	errorMessageMatch := "Invalid Attribute Value Match"
	paramsStrategy := map[string]interface{}{
		"name":                   name,
		"enabled_ldap":           "ldap2",
		"group_base_dn":          "CN=Users,DC=MyDomain,DC=com",
		"group_member_attribute": "uniqueMember",
		"sub_tree":               "true",
		"strategy":               "static",
	}
	t.Run(fmt.Sprintf("TestLdapGroup_StrategyCaseSensitive"), func(t *testing.T) {
		resource.Test(makeLdapGroupValidatorsTestCase(paramsStrategy, errorMessageMatch, t))
	})
}

func makeLdapGroupValidatorsTestCase(params map[string]interface{}, errorMessage string, t *testing.T) (*testing.T, resource.TestCase) {

	const ldapGroupSetting = `
	resource "artifactory_ldap_group_setting_v2" "{{ .name }}" {
		name = "{{ .name }}"
		enabled_ldap = "{{ .enabled_ldap }}"
		group_base_dn = "{{ .group_base_dn }}"
		group_name_attribute = "cn"
		group_member_attribute = "{{ .group_member_attribute }}"
		sub_tree = {{ .sub_tree }}
		force_attribute_search = false
		filter = "(objectClass=groupOfNames)"
		description_attribute = "description"
		strategy = "{{ .strategy }}"
	}
	`
	LdapSettingIncorrectDnPattern := utilsdk.ExecuteTemplate("TestLdap", ldapGroupSetting, params)

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,

		Steps: []resource.TestStep{
			{
				Config:      LdapSettingIncorrectDnPattern,
				ExpectError: regexp.MustCompile(errorMessage),
			},
		},
	}
}

func testAccLdapGroupSettingV2Destroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}
		resp, err := client.R().Head("access/api/v1/ldap/groups/" + rs.Primary.ID)

		if err != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				return nil
			}
			return err
		}

		return fmt.Errorf("error: LDAP Group Setting %s still exists", rs.Primary.ID)
	}
}
