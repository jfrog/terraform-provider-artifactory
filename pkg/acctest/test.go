package acctest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/provider"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v7/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"gopkg.in/yaml.v3"
)

const RtDefaultUser = "admin"

// Provider PreCheck(t) must be called before using this provider instance.
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

// PreCheck This function should be present in every acceptance test.
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
	return test.GetEnvVarWithFallback(t, "ARTIFACTORY_URL", "JFROG_URL")
}

type CheckFun func(id string, request *resty.Request) (*resty.Response, error)

func VerifyDeleted(id string, check CheckFun) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("error: Resource id [%s] not found", id)
		}

		if Provider == nil {
			return fmt.Errorf("provider is not initialized. Please PreCheck() is included in your acceptance test")
		}

		providerMeta := Provider.Meta().(util.ProvderMetadata)

		resp, err := check(rs.Primary.ID, providerMeta.Client.R())
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
	return repository.CheckRepo(id, request.AddRetryCondition(client.NeverRetry))
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

// CreateRepo Create a local repository with Xray indexing enabled. It will be used in the tests
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
		AddRetryCondition(client.RetryOnMergeError).
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
		AddRetryCondition(client.RetryOnMergeError).
		Delete("artifactory/api/repositories/" + repo)
	if errRepo != nil || response.StatusCode() != http.StatusOK {
		t.Logf("The repository %s doesn't exist", repo)
	}
}

// GetValidRandomDefaultRepoLayoutRef Usage of the function is strictly restricted to Test Cases
func GetValidRandomDefaultRepoLayoutRef() string {
	return test.RandSelect("simple-default", "bower-default", "composer-default", "conan-default", "go-default", "maven-2-default", "ivy-default", "npm-default", "nuget-default", "puppet-default", "sbt-default").(string)
}

// updateProxiesConfig is used by acctest.CreateProxy and acctest.DeleteProxy to interact with a proxy on Artifactory
var updateProxiesConfig = func(t *testing.T, proxyKey string, getProxiesBody func() []byte) {
	body := getProxiesBody()

	err := configuration.SendConfigurationPatch(body, Provider.Meta())
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
			Host:            "https://fake-proxy.org",
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
		// Return yaml to delete proxy
		proxiesConfig := fmt.Sprintf(`proxies:
  %s: ~`, proxyKey)
		return []byte(proxiesConfig)
	})
}

func GetTestResty(t *testing.T) *resty.Client {
	artifactoryUrl := GetArtifactoryUrl(t)
	restyClient, err := client.Build(artifactoryUrl, "")
	if err != nil {
		t.Fatal(err)
	}

	accessToken := test.GetEnvVarWithFallback(t, "ARTIFACTORY_ACCESS_TOKEN", "JFROG_ACCESS_TOKEN")
	api := os.Getenv("ARTIFACTORY_API_KEY")
	restyClient, err = client.AddAuth(restyClient, api, accessToken)
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

func DeleteUser(t *testing.T, name string) error {
	restyClient := GetTestResty(t)
	_, err := restyClient.R().Delete(user.UsersEndpointPath + name)

	return err
}

func CreateUserUpdatable(t *testing.T, name string, email string) {
	userObj := user.User{
		Name:                     name,
		Email:                    email,
		Password:                 "Lizard123!",
		Admin:                    false,
		ProfileUpdatable:         true,
		DisableUIAccess:          false,
		InternalPasswordDisabled: false,
		Groups:                   []string{"readers"},
	}

	restyClient := GetTestResty(t)
	_, err := restyClient.R().SetBody(userObj).Put(user.UsersEndpointPath + name)

	if err != nil {
		t.Fatal(err)
	}
}
