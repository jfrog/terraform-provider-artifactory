package artifactory

import (
	"encoding/xml"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const LdapGroupSettingTemplateFull = `
resource "artifactory_ldap_group_setting" "ldapgrouptest" {
	name = "ldapgrouptest"
	enabled_ldap = "ldaptest"
	group_base_dn = "ldaptest_dn"
	group_name_attribute = "cn"
	group_member_attribute = "uniqueMember"
	sub_tree = true
	filter = "(objectClass=groupOfNames)"
	description_attribute = "description"
	strategy = "STATIC"
}`

const LdapGroupSettingTemplateUpdate = `
resource "artifactory_ldap_group_setting" "ldapgrouptest" {
	name = "ldapgrouptest"
	enabled_ldap = "ldaptest"
	group_base_dn = "ldaptest1_dn"
	group_name_attribute = "cn"
	group_member_attribute = "uniqueMember"
	sub_tree = true
	filter = "(objectClass=groupOfNames)"
	description_attribute = "description1"
	strategy = "STATIC"
}`

func TestAccLdapGroupSetting_full(t *testing.T) {
	resource.Test(t, resource.TestCase{
		CheckDestroy:      testAccLdapGroupSettingDestroy("ldapgrouptest"),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: LdapGroupSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "enabled_ldap", "ldaptest"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "group_base_dn", "ldaptest_dn"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "group_name_attribute", "cn"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "group_member_attribute", "uniqueMember"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "sub_tree", "true"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "filter", "(objectClass=groupOfNames)"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "description_attribute", "description"),
				),
			},
			{
				Config: LdapGroupSettingTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "enabled_ldap", "ldaptest"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "group_base_dn", "ldaptest1_dn"),
					resource.TestCheckResourceAttr("artifactory_ldap_group_setting.ldapgrouptest", "description_attribute", "description1"),
				),
			},
		},
	})
}

func testAccLdapGroupSettingDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := testAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)

		_, ok := s.RootModule().Resources["artifactory_ldap_group_setting."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		ldapGroupConfigs := &XmlLdapGroupConfig{}

		response, err := client.R().Get("artifactory/api/system/configuration")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}

		err = xml.Unmarshal(response.Body(), &ldapGroupConfigs)
		if err != nil {
			return fmt.Errorf("failed to xml unmarshal ldap group settings during test destroy operation")
		}

		for _, iterLdapGroupSetting := range ldapGroupConfigs.LdapGroupSettings.LdapGroupSettingArr {
			if iterLdapGroupSetting.Name == id {
				return fmt.Errorf("error: LdapGroupSetting with name: " + id + " still exists.")
			}
		}
		return nil
	}
}
