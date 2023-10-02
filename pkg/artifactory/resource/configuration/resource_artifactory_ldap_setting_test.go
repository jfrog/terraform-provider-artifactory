package configuration_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/configuration"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
	"github.com/jfrog/terraform-provider-shared/validator"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
)

func TestAccLdapSetting_full(t *testing.T) {
	const LdapSettingTemplateFull = `
resource "artifactory_ldap_setting" "ldaptest" {
	key = "ldaptest"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=Peo *ple, uid={0}"
	email_attribute = "mail_attr"
	search_sub_tree = true
	search_filter = "(uid={0})"
	search_base = "ou=users|ou=people"
	manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
	manager_password = "testmgrpaswd"
}`

	const LdapSettingTemplateUpdate = `
resource "artifactory_ldap_setting" "ldaptest" {
	key = "ldaptest"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "mail_attr"
	search_sub_tree = true
	search_filter = "(uid={0})"
	search_base = "ou=users"
	manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
	manager_password = "testmgrpaswd"
}`
	var onOrAfterVersion7571 = func() (bool, error) {
		return acctest.CompareArtifactoryVersions(t, "7.57.1")
	}

	fqrn := "artifactory_ldap_setting.ldaptest"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccLdapSettingDestroy("ldaptest"),

		Steps: []resource.TestStep{
			{
				SkipFunc: onOrAfterVersion7571,
				Config:   LdapSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "ldap_url", "ldap://ldaptestldap"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", "ou=Peo *ple, uid={0}"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
					resource.TestCheckResourceAttr(fqrn, "search_sub_tree", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_filter", "(uid={0})"),
					resource.TestCheckResourceAttr(fqrn, "search_base", "ou=users|ou=people"),
				),
			},
			{
				Config: LdapSettingTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"manager_password"},
			},
		},
	})
}

func TestAccLdapSetting_importNotFound(t *testing.T) {
	config := `
		resource "artifactory_ldap_setting" "not-exist-test" {
			key = "not-exist-test"
			enabled = true
			ldap_url = "ldap://ldaptestldap"
			user_dn_pattern = "uid={0},ou=People"
			email_attribute = "mail_attr"
			search_sub_tree = true
			search_filter = "(uid={0})"
			search_base = "ou=users"
			manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
			manager_password = "testmgrpaswd"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        config,
				ResourceName:  "artifactory_ldap_setting.not-exist-test",
				ImportStateId: "not-exist-test",
				ImportState:   true,
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func TestAccLdapSetting_emailAttribute(t *testing.T) {
	const LdapSettingTemplateNoEmailAttr = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=People, uid={0}"
}`

	const LdapSettingTemplateEmailAttrBlank = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=People, uid={0}"
	email_attribute = ""
}`

	const LdapSettingTemplateEmailAttrUpd1 = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "mail"
}`

	const LdapSettingTemplateEmailAttrUpd2 = `
resource "artifactory_ldap_setting" "ldaptestemailattr" {
	key = "ldaptestemailattr"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "uid={0},ou=People"
	email_attribute = "mail_attr"
}`
	var onOrAfterVersion7571 = func() (bool, error) {
		return acctest.CompareArtifactoryVersions(t, "7.57.1")
	}

	fqrn := "artifactory_ldap_setting.ldaptestemailattr"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccLdapSettingDestroy("ldaptestemailattr"),

		Steps: []resource.TestStep{
			{
				SkipFunc: onOrAfterVersion7571,
				Config:   LdapSettingTemplateNoEmailAttr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail"),
				),
			},
			{
				Config: LdapSettingTemplateEmailAttrBlank,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail"),
				),
			},
			{
				Config: LdapSettingTemplateEmailAttrUpd1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail"),
				),
			},
			{
				Config: LdapSettingTemplateEmailAttrUpd2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState("ldaptestemailattr", "key"),
			},
		},
	})
}

func TestAccLdapSetting_user_dn_or_search_filter(t *testing.T) {
	const LdapSettingTemplateUserDnNoSearchFilter = `
resource "artifactory_ldap_setting" "ldaptestuserdnsearchfilter" {
	key = "ldaptestuserdnsearchfilter"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=People, uid={0}"
}`

	const LdapSettingTemplateSearchFilterNoUserDn = `
resource "artifactory_ldap_setting" "ldaptestuserdnsearchfilter" {
	key = "ldaptestuserdnsearchfilter"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	search_filter = "(uid={0})"
}`

	const LdapSettingTemplateUserDnAndSearchFilter = `
resource "artifactory_ldap_setting" "ldaptestuserdnsearchfilter" {
	key = "ldaptestuserdnsearchfilter"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
	user_dn_pattern = "ou=People, uid={0}"
    search_filter = "(uid={0})"
}`

	// Note: Artifactory REST API creates LDAP setting config even when both user_dn_pattern and search_filter are empty. In UI, User is prompted to specify values for either/both of these fields.
	const LdapSettingTemplateNoUserDnNoSearchFilter = `
resource "artifactory_ldap_setting" "ldaptestuserdnsearchfilter" {
	key = "ldaptestuserdnsearchfilter"
	enabled = true
	ldap_url = "ldap://ldaptestldap"
}`

	var onOrAfterVersion7571 = func() (bool, error) {
		return acctest.CompareArtifactoryVersions(t, "7.57.1")
	}

	fqrn := "artifactory_ldap_setting.ldaptestuserdnsearchfilter"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccLdapSettingDestroy("ldaptestuserdnsearchfilter"),

		Steps: []resource.TestStep{
			{
				SkipFunc: onOrAfterVersion7571,
				Config:   LdapSettingTemplateUserDnNoSearchFilter,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", "ou=People, uid={0}"),
				),
			},
			{
				Config: LdapSettingTemplateSearchFilterNoUserDn,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_filter", "(uid={0})"),
				),
			},
			{
				Config: LdapSettingTemplateUserDnAndSearchFilter,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", "ou=People, uid={0}"),
					resource.TestCheckResourceAttr(fqrn, "search_filter", "(uid={0})"),
				),
			},
			{
				Config: LdapSettingTemplateNoUserDnNoSearchFilter,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState("ldaptestuserdnsearchfilter", "key"),
			},
		},
	})
}

func testAccLdapSettingDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

		_, ok := s.RootModule().Resources["artifactory_ldap_setting."+id]
		if !ok {
			return fmt.Errorf("error: resource id [%s] not found", id)
		}
		ldapConfigs := &configuration.XmlLdapConfig{}

		response, err := client.R().SetResult(&ldapConfigs).Get("artifactory/api/system/configuration")
		if err != nil {
			return fmt.Errorf("error: failed to retrieve data from API: /artifactory/api/system/configuration during Read")
		}
		if response.IsError() {
			return fmt.Errorf("got error response for API: /artifactory/api/system/configuration request during Read")
		}

		for _, iterLdapSetting := range ldapConfigs.Security.LdapSettings.LdapSettingArr {
			if iterLdapSetting.Key == id {
				return fmt.Errorf("error: LdapSetting with key: " + id + " still exists.")
			}
		}
		return nil
	}
}
