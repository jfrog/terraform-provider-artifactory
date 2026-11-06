// Copyright (c) JFrog Ltd. (2026)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configuration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	testresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	artifactoryprovider "github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/provider"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/configuration"
	"github.com/jfrog/terraform-provider-shared/util"
)

// Use production schemas and registration, but configure only a local HTTP client.
type condaConfigTestProvider struct {
	provider.Provider
	client *resty.Client
}

// Configure supplies the loopback API client to the production resources.
func (p *condaConfigTestProvider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	resp.ResourceData = util.ProviderMetadata{Client: p.client, ProductId: "test"}
}

// TestAccCondaConfig checks enablement, disablement and import on an isolated live instance.
func TestAccCondaConfig(t *testing.T) {
	fqrn := "artifactory_conda_config.global"
	config := func(enabled bool) string {
		return fmt.Sprintf(`resource "artifactory_conda_config" "global" { shards_enabled = %t }`, enabled)
	}
	testresource.Test(t, testresource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []testresource.TestStep{
			{
				Config: config(true),
				Check: testresource.ComposeTestCheckFunc(
					testresource.TestCheckResourceAttr(fqrn, "id", "conda"),
					testresource.TestCheckResourceAttr(fqrn, "shards_enabled", "true"),
				),
			},
			{
				Config: config(false),
				Check:  testresource.TestCheckResourceAttr(fqrn, "shards_enabled", "false"),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateId:     "conda",
				ImportStateVerify: true,
			},
		},
	})
}

// TestCondaConfigTerraformLifecycle checks plans, convergence, import, drift and removal against a local API.
func TestCondaConfigTerraformLifecycle(t *testing.T) {
	var mu sync.Mutex
	puts := 0
	document := map[string]interface{}{
		"serverName":   "existing-server",
		"largeCounter": json.Number("9007199254740993"),
		"unknownList":  []interface{}{"keep", nil, true},
		"repositoriesConfig": map[string]interface{}{
			"conda": map[string]interface{}{"shardsEnabled": true, "futureOption": "keep"},
			"rpm":   map[string]interface{}{"existingSetting": map[string]interface{}{"enabled": true, "visibleOnUi": false}},
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/artifactory/api/system/usage" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if req.URL.Path != "/artifactory/ui/generalConfig" {
			http.NotFound(w, req)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case http.MethodGet:
			if err := json.NewEncoder(w).Encode(document); err != nil {
				t.Error(err)
			}
		case http.MethodPut:
			var received map[string]interface{}
			decoder := json.NewDecoder(req.Body)
			decoder.UseNumber()
			if err := decoder.Decode(&received); err != nil {
				t.Error(err)
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			conda := received["repositoriesConfig"].(map[string]interface{})["conda"].(map[string]interface{})
			desired := conda["shardsEnabled"]
			conda["shardsEnabled"] = document["repositoriesConfig"].(map[string]interface{})["conda"].(map[string]interface{})["shardsEnabled"]
			if !reflect.DeepEqual(received, document) {
				t.Errorf("PUT changed unrelated configuration: %#v", received)
				http.Error(w, "unrelated settings changed", http.StatusBadRequest)
				return
			}
			conda["shardsEnabled"] = desired
			document = received
			puts++
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	factory := func() (tfprotov6.ProviderServer, error) {
		return providerserver.NewProtocol6(&condaConfigTestProvider{
			Provider: artifactoryprovider.Framework()(),
			client:   resty.New().SetBaseURL(server.URL),
		})(), nil
	}
	config := func(enabled bool) string {
		return fmt.Sprintf(`resource "artifactory_conda_config" "global" { shards_enabled = %t }`, enabled)
	}
	fqrn := "artifactory_conda_config.global"
	testresource.UnitTest(t, testresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){"artifactory": factory},
		CheckDestroy: func(_ *terraform.State) error {
			mu.Lock()
			defer mu.Unlock()
			enabled := document["repositoriesConfig"].(map[string]interface{})["conda"].(map[string]interface{})["shardsEnabled"]
			if enabled != true || puts != 3 {
				return fmt.Errorf("expected destroy to leave sharding enabled with three preceding writes, got enabled=%v, writes=%d", enabled, puts)
			}
			return nil
		},
		Steps: []testresource.TestStep{
			{
				Config: config(false),
				ConfigPlanChecks: testresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionCreate),
				}},
				Check: testresource.ComposeTestCheckFunc(
					testresource.TestCheckResourceAttr(fqrn, "id", "conda"),
					testresource.TestCheckResourceAttr(fqrn, "shards_enabled", "false"),
				),
			},
			{
				Config: config(true),
				ConfigPlanChecks: testresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionUpdate),
				}},
				Check: testresource.TestCheckResourceAttr(fqrn, "shards_enabled", "true"),
			},
			{
				Config: config(true),
				ConfigPlanChecks: testresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				}},
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateId:     "conda",
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					mu.Lock()
					defer mu.Unlock()
					document["repositoriesConfig"].(map[string]interface{})["conda"].(map[string]interface{})["shardsEnabled"] = false
					document["anotherSetting"] = "updated outside Terraform"
				},
				Config: config(true),
				ConfigPlanChecks: testresource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionUpdate),
				}},
				Check: testresource.TestCheckResourceAttr(fqrn, "shards_enabled", "true"),
			},
		},
	})
}

// TestCondaConfigCreateResponses checks idempotency and safe failure handling for API and schema errors.
func TestCondaConfigCreateResponses(t *testing.T) {
	valid := `{"repositoriesConfig":{"conda":{"shardsEnabled":false}}}`
	for _, tc := range []struct {
		name         string
		body         string
		getStatus    int
		putStatus    int
		readbackFail bool
		errorText    string
		writes       int
	}{
		{name: "already enabled", body: `{"repositoriesConfig":{"conda":{"shardsEnabled":true}}}`},
		{name: "read forbidden", body: valid, getStatus: 403, errorText: "403"},
		{name: "endpoint absent", body: valid, getStatus: 404, errorText: "404"},
		{name: "read failure", body: valid, getStatus: 500, errorText: "500"},
		{name: "malformed JSON", body: "{", errorText: "decode"},
		{name: "missing repositories", body: `{}`, errorText: "repositoriesConfig"},
		{name: "missing Conda", body: `{"repositoriesConfig":{}}`, errorText: "Conda package settings"},
		{name: "missing flag", body: `{"repositoriesConfig":{"conda":{}}}`, errorText: "as a boolean"},
		{name: "null flag", body: `{"repositoriesConfig":{"conda":{"shardsEnabled":null}}}`, errorText: "as a boolean"},
		{name: "string flag", body: `{"repositoriesConfig":{"conda":{"shardsEnabled":"false"}}}`, errorText: "as a boolean"},
		{name: "write forbidden", body: valid, putStatus: 403, errorText: "403", writes: 1},
		{name: "write failure", body: valid, putStatus: 500, errorText: "500", writes: 1},
		{name: "write ignored", body: valid, errorText: "did not persist", writes: 1},
		{name: "readback failure", body: valid, readbackFail: true, errorText: "500", writes: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mu sync.Mutex
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/artifactory/api/system/usage" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				mu.Lock()
				defer mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				if req.Method == http.MethodPut {
					writes++
					status := tc.putStatus
					if status == 0 {
						status = http.StatusNoContent
					}
					w.WriteHeader(status)
					return
				}
				status := tc.getStatus
				if tc.readbackFail && writes > 0 {
					status = http.StatusInternalServerError
				}
				if status != 0 {
					w.WriteHeader(status)
				}
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()

			ctx := context.Background()
			r := configuration.NewCondaConfigResource()
			var schemaResp resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			var configureResp resource.ConfigureResponse
			r.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{
				ProviderData: util.ProviderMetadata{Client: resty.New().SetBaseURL(server.URL), ProductId: "test"},
			}, &configureResp)
			if configureResp.Diagnostics.HasError() {
				t.Fatal(configureResp.Diagnostics)
			}
			state := tfsdk.State{Schema: schemaResp.Schema}
			if diags := state.Set(ctx, &configuration.CondaConfigResourceModel{
				ID: types.StringUnknown(), ShardsEnabled: types.BoolValue(true),
			}); diags.HasError() {
				t.Fatal(diags)
			}
			createResp := resource.CreateResponse{State: state}
			r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: schemaResp.Schema, Raw: state.Raw}}, &createResp)
			if tc.errorText == "" {
				if createResp.Diagnostics.HasError() {
					t.Fatal(createResp.Diagnostics)
				}
			} else if !createResp.Diagnostics.HasError() || !strings.Contains(fmt.Sprint(createResp.Diagnostics), tc.errorText) {
				t.Fatalf("expected error containing %q, got %v", tc.errorText, createResp.Diagnostics)
			}
			mu.Lock()
			defer mu.Unlock()
			if writes != tc.writes {
				t.Fatalf("expected %d writes, got %d", tc.writes, writes)
			}
		})
	}
}

// TestCondaConfigImportRejectsOtherIDs rejects repository names as singleton import IDs.
func TestCondaConfigImportRejectsOtherIDs(t *testing.T) {
	var resp resource.ImportStateResponse
	configuration.NewCondaConfigResource().(resource.ResourceWithImportState).ImportState(
		context.Background(), resource.ImportStateRequest{ID: "repository-name"}, &resp,
	)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error for a non-global import ID")
	}
}
