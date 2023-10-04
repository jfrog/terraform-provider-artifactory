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

func TestAccLdapSettingV2_full_no_search(t *testing.T) {
	_, fqrn, key := testutil.MkNames("ldap-", "artifactory_ldap_setting_v2")

	const ldapSetting = `
	resource "artifactory_ldap_setting_v2" "{{ .key }}" {
		key = "{{ .key }}"
		enabled = true
		ldap_url = "ldap://ldaptestldap"
		user_dn_pattern = "{{ .user_dn_pattern }}"
		email_attribute = "mail_attr"
	}
	`

	params := map[string]interface{}{
		"key":             key,
		"user_dn_pattern": "uid={0},ou=People",
		"search_base":     "ou=users|ou=people",
	}
	LdapSettingTemplateFull := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccLdapSettingV2Destroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: LdapSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "ldap_url", "ldap://ldaptestldap"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", params["user_dn_pattern"].(string)),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
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

func TestAccLdapSettingV2_full_with_search(t *testing.T) {
	_, fqrn, key := testutil.MkNames("ldap-", "artifactory_ldap_setting_v2")

	const ldapSetting = `
	resource "artifactory_ldap_setting_v2" "{{ .key }}" {
		key = "{{ .key }}"
		enabled = true
		ldap_url = "ldap://ldaptestldap"
		user_dn_pattern = "{{ .user_dn_pattern }}"
		email_attribute = "mail_attr"
		search_sub_tree = true
		search_filter = "(uid={0})"
		search_base = "{{ .search_base }}"
		manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
		manager_password = "testmgrpaswd"
	}
	`

	params := map[string]interface{}{
		"key":             key,
		"user_dn_pattern": "uid={0},ou=People",
		"search_base":     "ou=users|ou=people",
	}
	LdapSettingTemplateFull := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, params)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccLdapSettingV2Destroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: LdapSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "ldap_url", "ldap://ldaptestldap"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", params["user_dn_pattern"].(string)),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
					resource.TestCheckResourceAttr(fqrn, "search_sub_tree", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_filter", "(uid={0})"),
					resource.TestCheckResourceAttr(fqrn, "search_base", params["search_base"].(string)),
				),
				ConfigPlanChecks: acctest.ConfigPlanChecks,
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

func TestAccLdapSettingV2_update(t *testing.T) {
	_, fqrn, key := testutil.MkNames("ldap-", "artifactory_ldap_setting_v2")

	const ldapSetting = `
	resource "artifactory_ldap_setting_v2" "{{ .key }}" {
		key = "{{ .key }}"
		enabled = true
		ldap_url = "ldap://ldaptestldap"
		user_dn_pattern = "{{ .user_dn_pattern }}"
		email_attribute = "mail_attr"
		search_sub_tree = true
		search_filter = "(uid={0})"
		search_base = "{{ .search_base }}"
		manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
		manager_password = "{{ .manager_password }}"
	}
	`

	params := map[string]interface{}{
		"key":              key,
		"user_dn_pattern":  "uid={0},ou=People",
		"search_base":      "ou=users|ou=people",
		"manager_password": "testmgrpaswd",
	}
	LdapSettingTemplateFull := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, params)

	paramsUpdate := map[string]interface{}{
		"key":              key,
		"user_dn_pattern":  "ou=People, uid={0}",
		"search_base":      "ou=users",
		"manager_password": "testmgrpaswd1",
	}
	LdapSettingTemplateUpdate := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, paramsUpdate)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccLdapSettingV2Destroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: LdapSettingTemplateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "ldap_url", "ldap://ldaptestldap"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", params["user_dn_pattern"].(string)),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
					resource.TestCheckResourceAttr(fqrn, "search_sub_tree", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_filter", "(uid={0})"),
					resource.TestCheckResourceAttr(fqrn, "search_base", params["search_base"].(string)),
				),
			},
			{
				Config: LdapSettingTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", paramsUpdate["user_dn_pattern"].(string)),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
					resource.TestCheckResourceAttr(fqrn, "search_base", paramsUpdate["search_base"].(string)),
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

func TestAccLdapSettingV2_validators(t *testing.T) {
	_, _, key := testutil.MkNames("ldap-", "artifactory_ldap_setting_v2")

	const ldapSetting = `
	resource "artifactory_ldap_setting_v2" "{{ .key }}" {
		key = "{{ .key }}"
		enabled = true
		ldap_url = "ldap://ldaptestldap"
		user_dn_pattern = "{{ .user_dn_pattern }}"
		email_attribute = "mail_attr"
		search_sub_tree = true
		search_filter = "{{ .search_filter }}"
		search_base = "{{ .search_base }}"
		manager_dn = "{{ .manager_dn }}"
		manager_password = "testmgrpaswd"
	}
	`

	paramsUserDnPattern := map[string]interface{}{
		"key":             key,
		"user_dn_pattern": "#!@#$%^&*()_+?><uid={0},ou=People",
		"search_filter":   "(uid={0})",
		"search_base":     "ou=users|ou=people",
		"manager_dn":      "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com",
	}
	LdapSettingIncorrectDnPattern := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, paramsUserDnPattern)

	paramsSearchFilter := map[string]interface{}{
		"key":             key,
		"user_dn_pattern": "uid={0},ou=People",
		"search_filter":   "#!@#$%^&*()_+?><(uid={0})",
		"search_base":     "ou=users|ou=people",
		"manager_dn":      "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com",
	}
	LdapSettingIncorrectSearchFilter := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, paramsSearchFilter)

	paramsSearchBase := map[string]interface{}{
		"key":             key,
		"user_dn_pattern": "uid={0},ou=People",
		"search_filter":   "(uid={0})",
		"search_base":     "#!@#$%^&*()_+?><|#!@#$%^&*()_+?><|#!@#$%^&*()_+?><|ou=users|ou=people",
		"manager_dn":      "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com",
	}
	LdapSettingIncorrectSearchBase := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, paramsSearchBase)

	paramsManagerDn := map[string]interface{}{
		"key":             key,
		"user_dn_pattern": "uid={0},ou=People",
		"search_filter":   "(uid={0})",
		"search_base":     "#!@#$%^&*()_+?><|#!@#$%^&*()_+?><|#!@#$%^&*()_+?><|ou=users|ou=people",
		"manager_dn":      "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com",
	}
	LdapSettingIncorrectanagerDn := utilsdk.ExecuteTemplate("TestLdap", ldapSetting, paramsManagerDn)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,

		Steps: []resource.TestStep{
			{
				Config:      LdapSettingIncorrectDnPattern,
				ExpectError: regexp.MustCompile("Incorrect Attribute Configuration"),
			},
			{
				Config:      LdapSettingIncorrectSearchFilter,
				ExpectError: regexp.MustCompile("Incorrect Attribute Configuration"),
			},
			{
				Config:      LdapSettingIncorrectSearchBase,
				ExpectError: regexp.MustCompile("Incorrect Attribute Configuration"),
			},
			{
				Config:      LdapSettingIncorrectanagerDn,
				ExpectError: regexp.MustCompile("Incorrect Attribute Configuration"),
			},
		},
	})
}

func TestAccLdapSettingV2_importNotFound(t *testing.T) {
	config := `
		resource "artifactory_ldap_setting_v2" "not-exist-test" {
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
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:        config,
				ResourceName:  "artifactory_ldap_setting_v2.not-exist-test",
				ImportStateId: "not-exist-test",
				ImportState:   true,
				ExpectError:   regexp.MustCompile("exit status 1"),
			},
		},
	})
}

func TestAccLdapSettingV2_emailAttribute(t *testing.T) {
	_, fqrn, key := testutil.MkNames("ldap-", "artifactory_ldap_setting_v2")
	params := map[string]interface{}{"key": key}

	const LdapSettingTemplateNoEmailAttr = `
		resource "artifactory_ldap_setting_v2" "{{ .key }}" {
			key = "{{ .key }}"
			enabled = true
			ldap_url = "ldap://ldaptestldap"
			user_dn_pattern = "ou=People, uid={0}"
		}`

	const LdapSettingTemplateEmailAttrBlank = `
		resource "artifactory_ldap_setting_v2" "{{ .key }}" {
			key = "{{ .key }}"
			enabled = true
			ldap_url = "ldap://ldaptestldap"
			user_dn_pattern = "ou=People, uid={0}"
			email_attribute = ""
		}`

	const LdapSettingTemplateEmailAttrUpd1 = `
		resource "artifactory_ldap_setting_v2" "{{ .key }}" {
			key = "{{ .key }}"
			enabled = true
			ldap_url = "ldap://ldaptestldap"
			user_dn_pattern = "uid={0},ou=People"
			email_attribute = "mail"
		}`

	const LdapSettingTemplateEmailAttrUpd2 = `
		resource "artifactory_ldap_setting_v2" "{{ .key }}" {
			key = "{{ .key }}"
			enabled = true
			ldap_url = "ldap://ldaptestldap"
			user_dn_pattern = "uid={0},ou=People"
			email_attribute = "mail_attr"
		}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccLdapSettingV2Destroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate("TestLdap", LdapSettingTemplateNoEmailAttr, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail"),
				),
			},
			{
				Config: utilsdk.ExecuteTemplate("TestLdap", LdapSettingTemplateEmailAttrBlank, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", ""),
				),
			},
			{
				Config: utilsdk.ExecuteTemplate("TestLdap", LdapSettingTemplateEmailAttrUpd1, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail"),
				),
			},
			{
				Config: utilsdk.ExecuteTemplate("TestLdap", LdapSettingTemplateEmailAttrUpd2, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "email_attribute", "mail_attr"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(key, "key"),
			},
		},
	})
}

func TestAccLdapSettingV2_user_dn_or_search_filter(t *testing.T) {
	_, fqrn, key := testutil.MkNames("ldap-", "artifactory_ldap_setting_v2")
	params := map[string]interface{}{"key": key}

	const LdapSettingTemplateUserDnNoSearchFilter = `
		resource "artifactory_ldap_setting_v2" "{{ .key }}" {
			key = "{{ .key }}"
			enabled = true
			ldap_url = "ldap://ldaptestldap"
			user_dn_pattern = "ou=People, uid={0}"
		}`

	const LdapSettingTemplateNoUserDn = `
		resource "artifactory_ldap_setting_v2" "{{ .key }}" {
			key = "{{ .key }}"
			enabled = true
			ldap_url = "ldap://ldaptestldap"
			search_sub_tree = true
			search_filter = "(uid={0})"
			search_base = "ou=users"
			manager_dn = "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"
			manager_password = "testmgrpaswd"
		}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccLdapSettingV2Destroy(fqrn),

		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate("TestLdap", LdapSettingTemplateUserDnNoSearchFilter, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "user_dn_pattern", "ou=People, uid={0}"),
				),
			},
			{
				Config: utilsdk.ExecuteTemplate("TestLdap", LdapSettingTemplateNoUserDn, params),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "ldap_url", "ldap://ldaptestldap"),
					resource.TestCheckResourceAttr(fqrn, "search_sub_tree", "true"),
					resource.TestCheckResourceAttr(fqrn, "search_filter", "(uid={0})"),
					resource.TestCheckResourceAttr(fqrn, "search_base", "ou=users"),
					resource.TestCheckResourceAttr(fqrn, "manager_dn", "CN=John Smith, OU=San Francisco,DC=am,DC=example,DC=com"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateCheck:        validator.CheckImportState(key, "key"),
				ImportStateVerifyIgnore: []string{"manager_password"},
			},
		},
	})
}

func testAccLdapSettingV2Destroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := acctest.Provider.Meta().(utilsdk.ProvderMetadata).Client

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}
		resp, err := client.R().Head("access/api/v1/ldap/settings/" + rs.Primary.ID)

		if err != nil {
			if resp != nil && resp.StatusCode() == http.StatusNotFound {
				return nil
			}
			return err
		}

		return fmt.Errorf("error: LDAP Settings %s still exists", rs.Primary.ID)
	}
}
