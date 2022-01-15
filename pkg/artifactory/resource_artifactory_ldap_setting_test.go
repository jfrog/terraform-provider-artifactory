package artifactory

import (
	"encoding/xml"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const LdapSettingTemplateFull = `
resource "artifactory_ldap_setting" "ldaptest" {
	key = "ldaptest"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "testldap@test.org"
	search_sub_tree = true
	search_filter = "(uid={0})"
	search_base = "ou=users"
	manager_dn = "testmgrdn"
	manager_password = "testmgrpaswd"
}`

const LdapSettingTemplateUpdate = `
resource "artifactory_ldap_setting" "ldaptest" {
	key = "ldaptest"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "testldapupdate@test.org"
	search_sub_tree = true
	search_filter = "(uid={0})"
	search_base = "ou=users"
	manager_dn = "testmgrdn"
	manager_password = "testmgrpaswd"
}`

func TestAccLdapSetting_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccLdapSettingDestroy("ldaptest"),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: LdapSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "ldap_url", "ldap://ldaptestldap"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "user_dn_pattern", "uid={0},ou=People"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "email_attribute", "testldap@test.org"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "search_sub_tree", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "search_filter", "(uid={0})"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "search_base", "ou=users"),
				),
			},
			{
				Config: LdapSettingTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "enabled", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_setting.ldaptest", "email_attribute", "testldapupdate@test.org"),
				),
			},
		},
	})
}

func testAccLdapSettingDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := testAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources["artifactory_ldap_setting."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		ldapConfigs := &XmlLdapConfig{}

		response, err := client.R().Get("artifactory/api/system/configuration")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from <base_url>/artifactory/api/system/configuration during Read")
		}

		err = xml.Unmarshal(response.Body(), &ldapConfigs)
		if err != nil {
			return fmt.Errorf("failed to xml unmarshal ldap settings during test destroy operation")
		}

		for _, iterLdapSetting := range ldapConfigs.LdapSettings.LdapSettingArr {
			if iterLdapSetting.Key == id {
				return fmt.Errorf("error: LdapSetting with key: " + id + " still exists.")
			}
		}
		return nil
	}
}
