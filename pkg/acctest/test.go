package acctest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraform2 "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/provider"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
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

// ProtoV5MuxProviderFactories is used to instantiate both SDK v2 and Framework providers
// during acceptance tests. Use it only if you need to combine resources from SDK v2 and the Framework in the same test.
var ProtoV5MuxProviderFactories map[string]func() (tfprotov5.ProviderServer, error)

var ProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"artifactory": providerserver.NewProtocol5WithError(provider.Framework()()),
}

func init() {
	Provider = provider.SdkV2()

	ProviderFactories = map[string]func() (*schema.Provider, error){
		"artifactory": func() (*schema.Provider, error) { return Provider, nil },
	}

	ProtoV5MuxProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
		"artifactory": func() (tfprotov5.ProviderServer, error) {
			ctx := context.Background()
			providers := []func() tfprotov5.ProviderServer{
				providerserver.NewProtocol5(provider.Framework()()), // terraform-plugin-framework provider
				Provider.GRPCProvider,                               // terraform-plugin-sdk provider
			}

			muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

			if err != nil {
				return nil, err
			}

			return muxServer.ProviderServer(), nil
		},
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
			t.Fatalf("failed to set custom base URL: %v", err)
		}

		configErr := Provider.Configure(context.Background(), (*terraform2.ResourceConfig)(terraform2.NewResourceConfigRaw(nil)))
		if configErr != nil && configErr.HasError() {
			t.Fatalf("failed to configure provider %v", configErr)
		}
	})
}

func GetArtifactoryUrl(t *testing.T) string {
	return testutil.GetEnvVarWithFallback(t, "JFROG_URL", "ARTIFACTORY_URL")
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

		providerMeta := Provider.Meta().(utilsdk.ProvderMetadata)

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
	return testutil.RandSelect("simple-default", "bower-default", "composer-default", "conan-default", "go-default", "maven-2-default", "ivy-default", "npm-default", "nuget-default", "puppet-default", "sbt-default").(string)
}

// updateProxiesConfig is used by acctest.CreateProxy and acctest.DeleteProxy to interact with a proxy on Artifactory
var updateProxiesConfig = func(t *testing.T, proxyKey string, getProxiesBody func() []byte) {
	body := getProxiesBody()
	restyClient := GetTestResty(t)
	metadata := utilsdk.ProvderMetadata{Client: restyClient}
	err := configuration.SendConfigurationPatch(body, metadata)
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

	accessToken := testutil.GetEnvVarWithFallback(t, "JFROG_ACCESS_TOKEN", "ARTIFACTORY_ACCESS_TOKEN")
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

func CompareArtifactoryVersions(t *testing.T, instanceVersions string) (bool, error) {

	fixedVersion, err := version.NewVersion(instanceVersions)
	if err != nil {
		return false, err
	}

	meta := Provider.Meta().(utilsdk.ProvderMetadata)
	runtimeVersion, err := version.NewVersion(meta.ArtifactoryVersion)
	if err != nil {
		return false, err
	}

	skipTest := runtimeVersion.GreaterThanOrEqual(fixedVersion)
	if skipTest {
		t.Skip(fmt.Printf("Test skip because: runtime version %s is same or later than %s\n", runtimeVersion.String(), fixedVersion.String()))
	}
	return skipTest, nil
}

var ConfigPlanChecks = resource.ConfigPlanChecks{
	PostApplyPreRefresh: []plancheck.PlanCheck{
		DebugPlan("PostApplyPreRefresh"),
	},
	PostApplyPostRefresh: []plancheck.PlanCheck{
		DebugPlan("PostApplyPostRefresh"),
	},
}

var _ plancheck.PlanCheck = PlanCheck{}

type PlanCheck struct {
	Stage string
}

func (p PlanCheck) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	var err error

	rc, err := json.Marshal(req.Plan.ResourceChanges[0])
	if err != nil {
		resp.Error = err
		return
	}

	pv, err := json.Marshal(req.Plan.PlannedValues)
	if err != nil {
		resp.Error = err
		return
	}

	ps, err := json.Marshal(req.Plan.PriorState)
	if err != nil {
		resp.Error = err
		return
	}

	rd, err := json.Marshal(req.Plan.ResourceDrift)
	if err != nil {
		resp.Error = err
		return
	}

	tflog.Debug(ctx, "CheckPlan", map[string]interface{}{
		"stage":                                  p.Stage,
		"req.Plan.ResourceChanges.ResourceDrift": string(rd),
		"req.Plan.ResourceChanges":               string(rc),
		"req.Plan.PlannedValues":                 string(pv),
		"req.Plan.PriorState":                    string(ps),
	})

	if len(req.Plan.ResourceDrift) > 0 {
		resp.Error = fmt.Errorf("expected empty plan, but has resouce drifts(s): %v", req.Plan.ResourceDrift)
		return
	}

	var errStrings []string
	for _, rc := range req.Plan.ResourceChanges {
		if !rc.Change.Actions.NoOp() {
			errStrings = append(errStrings, fmt.Sprintf("expected empty plan, but %s has planned action(s): %v", rc.Address, rc.Change.Actions))
		}
	}

	if err != nil {
		resp.Error = fmt.Errorf(strings.Join(errStrings, "\n"))
		return
	}
}

func DebugPlan(stage string) plancheck.PlanCheck {
	return PlanCheck{
		Stage: stage,
	}
}
