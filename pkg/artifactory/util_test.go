package artifactory

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func fmtMapToHcl(fields map[string]interface{}) string {
	var allPairs []string
	max := float64(0)
	for key := range fields {
		max = math.Max(max, float64(len(key)))
	}
	for key, value := range fields {
		hcl := toHclFormat(value)
		format := toHclFormatString(3, int(max), value)
		allPairs = append(allPairs, fmt.Sprintf(format, key, hcl))
	}

	return strings.Join(allPairs, "\n")
}
func toHclFormatString(tabs, max int, value interface{}) string {
	prefix := ""
	suffix := ""
	delimeter := "="
	if reflect.TypeOf(value).Kind() == reflect.Map {
		delimeter = ""
		prefix = "{"
		suffix = "}"
	}
	return fmt.Sprintf("%s%%-%ds %s %s%s%s", strings.Repeat("\t", tabs), max, delimeter, prefix, "%s", suffix)
}
func mapToTestChecks(fqrn string, fields map[string]interface{}) []resource.TestCheckFunc {
	var result []resource.TestCheckFunc
	for key, value := range fields {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Slice:
			for i, lv := range value.([]interface{}) {
				result = append(result, resource.TestCheckResourceAttr(
					fqrn,
					fmt.Sprintf("%s.%d", key, i),
					fmt.Sprintf("%v", lv),
				))
			}
		case reflect.Map:
			// this also gets generated, but it's value is '1', which is also the size. So, I don't know
			// what it means
			// content_synchronisation.0.%
			resource.TestCheckResourceAttr(
				fqrn,
				fmt.Sprintf("%s.#", key),
				fmt.Sprintf("%d", len(value.(map[string]interface{}))),
			)
		default:
			result = append(result, resource.TestCheckResourceAttr(fqrn, key, fmt.Sprintf(`%v`, value)))
		}
	}
	return result
}
func toHclFormat(thing interface{}) string {
	switch thing.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, thing.(string))
	case []interface{}:
		var result []string
		for _, e := range thing.([]interface{}) {
			result = append(result, toHclFormat(e))
		}
		return fmt.Sprintf("[%s]", strings.Join(result, ","))
	case map[string]interface{}:
		return fmt.Sprintf("\n\t%s\n\t\t\t\t", fmtMapToHcl(thing.(map[string]interface{})))
	default:
		return fmt.Sprintf("%v", thing)
	}
}

type CheckFun func(id string, request *resty.Request) (*resty.Response, error)

func verifyDeleted(id string, check CheckFun) func(*terraform.State) error {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("error: Resource id [%s] not found", id)
		}
		provider, _ := testAccProviders["artifactory"]()
		client := provider.Meta().(*resty.Client)
		resp, err := check(rs.Primary.ID, client.R())
		if err != nil {
			if resp != nil {
				switch resp.StatusCode() {
				case http.StatusNotFound, http.StatusBadRequest:
					return nil
				}
			}
			return err
		}
		return fmt.Errorf("error: %s still exists", rs.Primary.ID)
	}
}

func testCheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	return checkRepo(id, request.AddRetryCondition(neverRetry))
}

func createProject(t *testing.T, projectKey string) {
	type AdminPrivileges struct {
		ManageMembers   bool `json:"manage_members"`
		ManageResources bool `json:"manage_resources"`
		IndexResources  bool `json:"index_resources"`
	}

	type Project struct {
		Key             string          `json:"project_key"`
		DisplayName     string          `json:"display_name"`
		Description     string          `json:"description"`
		AdminPrivileges AdminPrivileges `json:"admin_privileges"`
	}

	restyClient := getTestResty(t)

	project := Project{
		Key:         projectKey,
		DisplayName: projectKey,
		Description: fmt.Sprintf("%s description", projectKey),
		AdminPrivileges: AdminPrivileges{
			ManageMembers:   true,
			ManageResources: true,
			IndexResources:  true,
		},
	}

	_, err := restyClient.R().SetBody(project).Post("/access/api/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
}

func deleteProject(t *testing.T, projectKey string) {
	restyClient := getTestResty(t)
	_, err := restyClient.R().Delete("/access/api/v1/projects/" + projectKey)
	if err != nil {
		t.Fatal(err)
	}
}

// Create a local repository with Xray indexing enabled. It will be used in the tests
func testAccCreateRepos(t *testing.T, repo string, rclass string, packageType string,
	handleReleases bool, handleSnapshots bool) {
	restyClient := getTestResty(t)

	type Repository struct {
		Rclass                  string `json:"rclass"`
		PackageType             string `json:"packageType"`
		HandleReleases          bool   `json:"handleReleases"`
		HandleSnapshots         bool   `json:"handleSnapshots"`
		SnapshotVersionBehavior string `json:"snapshotVersionBehavior"`
		XrayIndex               bool   `json:"xrayIndex"`
	}

	repository := Repository{}
	repository.Rclass = rclass
	repository.PackageType = packageType
	repository.HandleReleases = handleReleases
	repository.HandleSnapshots = handleSnapshots
	repository.SnapshotVersionBehavior = "unique"
	repository.XrayIndex = true
	response, errRepo := restyClient.R().SetBody(repository).Put("artifactory/api/repositories/" + repo)
	//Artifactory can return 400 for several reasons, this is why we are checking the response body
	repoExists := strings.Contains(fmt.Sprint(errRepo), "Case insensitive repository key already exists")
	if !repoExists && response.StatusCode() != http.StatusOK {
		t.Error(errRepo)
	}
}

func testAccDeleteRepo(t *testing.T, repo string) {
	restyClient := getTestResty(t)

	response, errRepo := restyClient.R().Delete("artifactory/api/repositories/" + repo)
	if errRepo != nil || response.StatusCode() != http.StatusOK {
		t.Logf("The repository %s doesn't exist", repo)
	}
}

//Usage of the function is strictly restricted to Test Cases
func getValidRandomDefaultRepoLayoutRef() string {
	return randSelect("simple-default", "bower-default", "composer-default", "conan-default", "go-default", "maven-2-default", "ivy-default", "npm-default", "nuget-default", "puppet-default", "sbt-default").(string)
}

// updateProxiesConfig is used by createProxy and deleteProxy to interact with a proxy on Artifactory
var updateProxiesConfig = func(t *testing.T, proxyKey string, getProxiesBody func() []byte) {
	body := getProxiesBody()
	restyClient := getTestResty(t)

	err := sendConfigurationPatch(body, restyClient)
	if err != nil {
		t.Fatal(err)
	}
}

// createProxy creates a new proxy on Artifactory with the given key
var createProxy = func(t *testing.T, proxyKey string) {
	type proxy struct {
		Key             string `yaml:"key"`
		Host            string `yaml:"host"`
		Port            int    `yaml:"port"`
		PlatformDefault bool   `yaml:"platformDefault"`
	}

	updateProxiesConfig(t, proxyKey, func() []byte {
		testProxy := proxy{
			Key:             proxyKey,
			Host:            "http://fake-proxy.org",
			Port:            8080,
			PlatformDefault: false,
		}

		constructBody := map[string][]proxy{
			"proxies": {testProxy},
		}

		body, err := yaml.Marshal(&constructBody)
		if err != nil {
			t.Errorf("failed to marshal proxies settings during Update")
		}

		return body
	})
}

// createProxy deletes an existing proxy on Artifactory with the given key
var deleteProxy = func(t *testing.T, proxyKey string) {
	updateProxiesConfig(t, proxyKey, func() []byte {
		// Return empty yaml to clean up all proxies
		return []byte(`proxies: ~`)
	})
}

func addTestCertificate(t *testing.T, certificateAlias string) {
	restyClient := getTestResty(t)

	certFileBytes, err := ioutil.ReadFile("../../samples/cert.pem")
	if err != nil {
		t.Fatal(err)
	}

	_, err = restyClient.R().
		SetBody(string(certFileBytes)).
		SetContentLength(true).
		Post(fmt.Sprintf("%s%s", certificateEndpoint, certificateAlias))
	if err != nil {
		t.Fatal(err)
	}
}

func deleteTestCertificate(t *testing.T, certificateAlias string) {
	restyClient := getTestResty(t)

	_, err := restyClient.R().
		Delete(fmt.Sprintf("%s%s", certificateEndpoint, certificateAlias))
	if err != nil {
		t.Fatal(err)
	}
}
