package acctest

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"
	"text/template"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/provider"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
	"gopkg.in/yaml.v2"
)

const RtDefaultUser = "admin"

// PreCheck(t) must be called before using this provider instance.
var Provider *schema.Provider
var ProviderFactories map[string]func() (*schema.Provider, error)

// testAccProviderConfigure ensures Provider is only configured once
//
// The PreCheck(t) function is invoked for every test and this prevents
// extraneous reconfiguration to the same values each time. However, this does
// not prevent reconfiguration that may happen should the address of
// Provider be errantly reused in ProviderFactories.
var testAccProviderConfigure sync.Once

func init() {
	Provider = provider.Provider()

	ProviderFactories = map[string]func() (*schema.Provider, error){
		"artifactory": func() (*schema.Provider, error) { return provider.Provider(), nil },
	}
}

// This PreCheck function should be present in every acceptance test.
func PreCheck(t *testing.T) {
	// Since we are outside the scope of the Terraform configuration we must
	// call Configure() to properly initialize the provider configuration.
	testAccProviderConfigure.Do(func() {
		restyClient := GetTestResty(t)

		artifactoryUrl := GetArtifactoryUrl(t)
		// Set custom base URL so repos that relies on it will work
		// https://www.jfrog.com/confluence/display/JFROG/Artifactory+REST+API#ArtifactoryRESTAPI-UpdateCustomURLBase
		_, err := restyClient.R().
			SetBody(artifactoryUrl).
			SetHeader("Content-Type", "text/plain").
			Put("/artifactory/api/system/configuration/baseUrl")
		if err != nil {
			t.Fatalf("Failed to set custom base URL: %v", err)
		}

		configErr := Provider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
		if configErr != nil {
			t.Fatalf("Failed to configure provider %v", configErr)
		}
	})
}

func GetArtifactoryUrl(t *testing.T) string {
	var ok bool
	var artifactoryUrl string
	if artifactoryUrl, ok = os.LookupEnv("ARTIFACTORY_URL"); !ok {
		if artifactoryUrl, ok = os.LookupEnv("JFROG_URL"); !ok {
			t.Fatal("ARTIFACTORY_URL or JFROG_URL must be set for acceptance tests")
			return ""
		}
	}

	return artifactoryUrl
}

func FmtMapToHcl(fields map[string]interface{}) string {
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

func MapToTestChecks(fqrn string, fields map[string]interface{}) []resource.TestCheckFunc {
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
		return fmt.Sprintf("\n\t%s\n\t\t\t\t", FmtMapToHcl(thing.(map[string]interface{})))
	default:
		return fmt.Sprintf("%v", thing)
	}
}

func ExecuteTemplate(name, temp string, fields interface{}) string {
	var tpl bytes.Buffer
	if err := template.Must(template.New(name).Parse(temp)).Execute(&tpl, fields); err != nil {
		panic(err)
	}

	return tpl.String()
}

func MergeMaps(schemata ...map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, schma := range schemata {
		for k, v := range schma {
			result[k] = v
		}
	}
	return result
}

func CopyInterfaceMap(source map[string]interface{}, target map[string]interface{}) map[string]interface{} {
	for k, v := range source {
		target[k] = v
	}
	return target
}

func MkNames(name, resource string) (int, string, string) {
	id := utils.RandomInt()
	n := fmt.Sprintf("%s%d", name, id)
	return id, fmt.Sprintf("%s.%s", resource, n), n
}

type CheckFun func(id string, request *resty.Request) (*resty.Response, error)

func VerifyDeleted(id string, check CheckFun) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: Resource id [%s] not found", id)
		}

		if Provider == nil {
			return fmt.Errorf("Provider is not initialized. Please PreCheck() is included in your acceptance test.")
		}

		client := Provider.Meta().(*resty.Client)

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

func CheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	return repository.CheckRepo(id, request.AddRetryCondition(utils.NeverRetry))
}

func CreateProject(t *testing.T, projectKey string) {
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

	restyClient := GetTestResty(t)

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

	_, err := restyClient.R().
		SetBody(project).
		Post("/access/api/v1/projects")
	if err != nil {
		t.Fatal(err)
	}
}

func DeleteProject(t *testing.T, projectKey string) {
	restyClient := GetTestResty(t)
	_, err := restyClient.R().Delete("/access/api/v1/projects/" + projectKey)
	if err != nil {
		t.Fatal(err)
	}
}

// Create a local repository with Xray indexing enabled. It will be used in the tests
func CreateRepo(t *testing.T, repo string, rclass string, packageType string,
	handleReleases bool, handleSnapshots bool) {
	restyClient := GetTestResty(t)

	type Repository struct {
		Rclass                  string `json:"rclass"`
		PackageType             string `json:"packageType"`
		HandleReleases          bool   `json:"handleReleases"`
		HandleSnapshots         bool   `json:"handleSnapshots"`
		SnapshotVersionBehavior string `json:"snapshotVersionBehavior"`
		XrayIndex               bool   `json:"xrayIndex"`
	}

	r := Repository{}
	r.Rclass = rclass
	r.PackageType = packageType
	r.HandleReleases = handleReleases
	r.HandleSnapshots = handleSnapshots
	r.SnapshotVersionBehavior = "unique"
	r.XrayIndex = true
	response, errRepo := restyClient.R().
		SetBody(r).
		AddRetryCondition(utils.RetryOnMergeError).
		Put("artifactory/api/repositories/" + repo)
	//Artifactory can return 400 for several reasons, this is why we are checking the response body
	repoExists := strings.Contains(fmt.Sprint(errRepo), "Case insensitive repository key already exists")
	if !repoExists && response.StatusCode() != http.StatusOK {
		t.Error(errRepo)
	}
}

func DeleteRepo(t *testing.T, repo string) {
	restyClient := GetTestResty(t)

	response, errRepo := restyClient.R().
		AddRetryCondition(utils.RetryOnMergeError).
		Delete("artifactory/api/repositories/" + repo)
	if errRepo != nil || response.StatusCode() != http.StatusOK {
		t.Logf("The repository %s doesn't exist", repo)
	}
}

//Usage of the function is strictly restricted to Test Cases
func GetValidRandomDefaultRepoLayoutRef() string {
	return utils.RandSelect("simple-default", "bower-default", "composer-default", "conan-default", "go-default", "maven-2-default", "ivy-default", "npm-default", "nuget-default", "puppet-default", "sbt-default").(string)
}

// updateProxiesConfig is used by acctest.CreateProxy and acctest.DeleteProxy to interact with a proxy on Artifactory
var updateProxiesConfig = func(t *testing.T, proxyKey string, getProxiesBody func() []byte) {
	body := getProxiesBody()
	restyClient := GetTestResty(t)

	err := configuration.SendConfigurationPatch(body, restyClient)
	if err != nil {
		t.Fatal(err)
	}
}

// CreateProxy creates a new proxy on Artifactory with the given key
func CreateProxy(t *testing.T, proxyKey string) {
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

// DeleteProxy deletes an existing proxy on Artifactory with the given key
func DeleteProxy(t *testing.T, proxyKey string) {
	updateProxiesConfig(t, proxyKey, func() []byte {
		// Return empty yaml to clean up all proxies
		return []byte(`proxies: ~`)
	})
}

func GetTestResty(t *testing.T) *resty.Client {
	var ok bool
	var artifactoryUrl string
	if artifactoryUrl, ok = os.LookupEnv("ARTIFACTORY_URL"); !ok {
		if artifactoryUrl, ok = os.LookupEnv("JFROG_URL"); !ok {
			t.Fatal("ARTIFACTORY_URL or JFROG_URL must be set for acceptance tests")
		}
	}
	restyClient, err := utils.BuildResty(artifactoryUrl, "")
	if err != nil {
		t.Fatal(err)
	}

	var accessToken string
	if accessToken, ok = os.LookupEnv("ARTIFACTORY_ACCESS_TOKEN"); !ok {
		if accessToken, ok = os.LookupEnv("JFROG_ACCESS_TOKEN"); !ok {
			t.Fatal("ARTIFACTORY_ACCESS_TOKEN or JFROG_ACCESS_TOKEN must be set for acceptance tests")
		}
	}
	api := os.Getenv("ARTIFACTORY_API_KEY")
	restyClient, err = utils.AddAuthToResty(restyClient, api, accessToken)
	if err != nil {
		t.Fatal(err)
	}
	return restyClient
}

func CompositeCheckDestroy(funcs ...func(state *terraform.State) error) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		var errors []error
		for _, f := range funcs {
			err := f(state)
			if err != nil {
				errors = append(errors, err)
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("%q", errors)
		}
		return nil
	}
}
