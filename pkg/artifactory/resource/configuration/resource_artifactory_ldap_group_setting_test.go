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
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccLdapGroupSetting_full(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	const LdapGroupSettingTemplateFull = `
resource "artifactory_ldap_group_setting" "ldapgrouptest" {
	name = "ldapgrouptest"
	ldap_setting_key = "ldaptest"
	group_base_dn = "CN=Users,DC=MyDomain,DC=com"
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
	ldap_setting_key = "ldaptest"
	group_base_dn = "CN=Users,DC=MyDomain,DC=com"
	group_name_attribute = "cn"
	group_member_attribute = "uniqueMember"
	sub_tree = true
	filter = "(objectClass=groupOfNames)"
	description_attribute = "description1"
	strategy = "STATIC"
}`

	fqrn := "artifactory_ldap_group_setting.ldapgrouptest"

	var onOrAfterVersion7571 = func() (bool, error) {
		return acctest.CompareArtifactoryVersions(t, "7.57.1")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccLdapGroupSettingDestroy("ldapgrouptest"),

		Steps: []resource.TestStep{
			{
				SkipFunc: onOrAfterVersion7571,
				Config:   LdapGroupSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "ldap_setting_key", "ldaptest"),
					resource.TestCheckResourceAttr(fqrn, "group_base_dn", "CN=Users,DC=MyDomain,DC=com"),
					resource.TestCheckResourceAttr(fqrn, "group_name_attribute", "cn"),
					resource.TestCheckResourceAttr(fqrn, "group_member_attribute", "uniqueMember"),
					resource.TestCheckResourceAttr(fqrn, "sub_tree", "true"),
					resource.TestCheckResourceAttr(fqrn, "filter", "(objectClass=groupOfNames)"),
					resource.TestCheckResourceAttr(fqrn, "description_attribute", "description"),
				),
			},
			{
				Config: LdapGroupSettingTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "ldap_setting_key", "ldaptest"),
					resource.TestCheckResourceAttr(fqrn, "group_base_dn", "CN=Users,DC=MyDomain,DC=com"),
					resource.TestCheckResourceAttr(fqrn, "description_attribute", "description1"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState("ldapgrouptest", "name"),
			},
		},
	})
}

func TestAccLdapGroupSetting_importNotFound(t *testing.T) {
	jfrogURL := os.Getenv("JFROG_URL")
	if strings.HasSuffix(jfrogURL, "jfrog.io") {
		t.Skipf("env var JFROG_URL '%s' is a cloud instance.", jfrogURL)
	}

	var onOrAfterVersion7571 = func() (bool, error) {
		return acctest.CompareArtifactoryVersions(t, "7.57.1")
	}

	config := `
		resource "artifactory_ldap_group_setting" "not-exist-test" {
			name = "not-exist-test"
			ldap_setting_key = "ldaptest"
			group_base_dn = "CN=Users,DC=MyDomain,DC=com"
			group_name_attribute = "cn"
			group_member_attribute = "uniqueMember"
			sub_tree = true
			filter = "(objectClass=groupOfNames)"
			description_attribute = "description"
			strategy = "STATIC"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				SkipFunc:      onOrAfterVersion7571,
				Config:        config,
				ResourceName:  "artifactory_ldap_group_setting.not-exist-test",
				ImportStateId: "not-exist-test",
				ImportState:   true,
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func testAccLdapGroupSettingDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(util.ProviderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_ldap_group_setting."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		ldapGroupConfigs := &configuration.XmlLdapGroupConfig{}

		response, err := client.R().SetResult(&ldapGroupConfigs).Get(configuration.ConfigurationEndpoint)
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		for _, iterLdapGroupSetting := range ldapGroupConfigs.Security.LdapGroupSettings.LdapGroupSettingArr {
			if iterLdapGroupSetting.Name == id {
				return fmt.Errorf("error: LdapGroupSetting with name: " + id + " still exists.")
			}
		}
		return nil
	}
}
