package artifact_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccItemProperties_repo_only(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-item-properties-", "artifactory_item_properties")
	_, _, repoName := testutil.MkNames("test-generic-local", "artifactory_local_generic_repository")

	temp := `
	resource "artifactory_local_generic_repository" "{{ .repoName }}" {
		key = "{{ .repoName }}"
	}

	resource "artifactory_item_properties" "{{ .name }}" {
		repo_key = artifactory_local_generic_repository.{{ .repoName }}.key
		properties = {
			"key1" = ["value1", "value2"],
			"key2" = ["value3", "value4"]
		}
		is_recursive = true
	}`

	testData := map[string]string{
		"name":     name,
		"repoName": repoName,
	}
	config := util.ExecuteTemplate(name, temp, testData)

	updatedTemp := `
	resource "artifactory_local_generic_repository" "{{ .repoName }}" {
		key = "{{ .repoName }}"
	}

	resource "artifactory_item_properties" "{{ .name }}" {
		repo_key = artifactory_local_generic_repository.{{ .repoName }}.key
		properties = {
			"key1" = ["value1"]
			"key3" = ["value5"]
		}
		is_recursive = false
	}`
	updatedConfig := util.ExecuteTemplate(name, updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", repoName),
					resource.TestCheckNoResourceAttr(fqrn, "item_path"),
					resource.TestCheckResourceAttr(fqrn, "properties.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "properties.key1.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key1.*", "value1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key1.*", "value2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key2.*", "value3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key2.*", "value4"),
					resource.TestCheckResourceAttr(fqrn, "is_recursive", "true"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", repoName),
					resource.TestCheckNoResourceAttr(fqrn, "item_path"),
					resource.TestCheckResourceAttr(fqrn, "properties.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "properties.key1.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key1.*", "value1"),
					resource.TestCheckResourceAttr(fqrn, "properties.key3.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key3.*", "value5"),
					resource.TestCheckResourceAttr(fqrn, "is_recursive", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        testData["repoName"],
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "repo_key",
				ImportStateVerifyIgnore:              []string{"is_recursive"},
			},
		},
	})
}

func createRepoPath(t *testing.T, repoKey, path string) {
	restyClient := acctest.GetTestResty(t)
	response, err := restyClient.R().
		SetRawPathParams(map[string]string{
			"repo_key": repoKey,
			"path":     path,
		}).
		Put("artifactory/{repo_key}/{path}")

	if err != nil {
		t.Error(err)
	}

	if response.IsError() {
		t.Error(response.String())
	}
}

func TestAccItemProperties_repo_with_path(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-item-properties-", "artifactory_item_properties")
	_, _, repoName := testutil.MkNames("test-generic-local", "artifactory_local_generic_repository")

	temp := `
	resource "artifactory_item_properties" "{{ .name }}" {
		repo_key = "{{ .repoName }}"
		item_path = "foo/bar"
		properties = {
			"key1" = ["value1", "value2"],
			"key2" = ["value3", "value4"]
		}
		is_recursive = true
	}`

	testData := map[string]string{
		"name":     name,
		"repoName": repoName,
	}
	config := util.ExecuteTemplate(name, temp, testData)

	updatedTemp := `
	resource "artifactory_item_properties" "{{ .name }}" {
		repo_key = "{{ .repoName }}"
		item_path = "foo/bar"
		properties = {
			"key1" = ["value1"]
			"key3" = ["value5"]
		}
		is_recursive = false
	}`
	updatedConfig := util.ExecuteTemplate(name, updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)

			acctest.CreateRepo(t, repoName, "local", "generic", false, false)
			createRepoPath(t, repoName, "foo/bar")
		},
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			acctest.DeleteRepo(t, repoName)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", repoName),
					resource.TestCheckResourceAttr(fqrn, "item_path", "foo/bar"),
					resource.TestCheckResourceAttr(fqrn, "properties.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "properties.key1.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key1.*", "value1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key1.*", "value2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key2.*", "value3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key2.*", "value4"),
					resource.TestCheckResourceAttr(fqrn, "is_recursive", "true"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "repo_key", repoName),
					resource.TestCheckResourceAttr(fqrn, "item_path", "foo/bar"),
					resource.TestCheckResourceAttr(fqrn, "properties.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "properties.key1.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key1.*", "value1"),
					resource.TestCheckResourceAttr(fqrn, "properties.key3.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "properties.key3.*", "value5"),
					resource.TestCheckResourceAttr(fqrn, "is_recursive", "false"),
				),
			},
			{
				ResourceName:                         fqrn,
				ImportState:                          true,
				ImportStateId:                        fmt.Sprintf("%s:foo/bar", testData["repoName"]),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "repo_key",
				ImportStateVerifyIgnore:              []string{"is_recursive"},
			},
		},
	})
}

func TestAccItemProperties_invalid_property_key_character(t *testing.T) {
	invalidChars := []string{")", "(", "}", "{", "]", "[", "*", "+", "^", "$", "/", "~", "`", "!", "@", "#", "%", "&", "<", ">", ";", "=", ",", "±", "§", " "}
	for _, invalidChar := range invalidChars {
		t.Run(invalidChar, func(t *testing.T) {
			resource.Test(testInvalidKey(invalidChar, t))
		})
	}
}

func testInvalidKey(invalidChar string, t *testing.T) (*testing.T, resource.TestCase) {
	_, _, name := testutil.MkNames("test-item-properties-", "artifactory_item_properties")
	_, _, repoName := testutil.MkNames("test-generic-local", "artifactory_local_generic_repository")

	temp := `
	resource "artifactory_local_generic_repository" "{{ .repoName }}" {
		key = "{{ .repoName }}"
	}

	resource "artifactory_item_properties" "{{ .name }}" {
		repo_key = artifactory_local_generic_repository.{{ .repoName }}.key
		properties = {
			"invalid_{{ .invalid_char }}_key" = ["value1", "value2"],
		}
	}`

	testData := map[string]string{
		"name":         name,
		"repoName":     repoName,
		"invalid_char": invalidChar,
	}
	config := util.ExecuteTemplate(name, temp, testData)

	return t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`.*must not contain the following special\n.*characters.*`),
			},
		},
	}
}
