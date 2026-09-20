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

package configuration

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPackageCleanupPolicyFromAPIModelPreservesNullCronExpression(t *testing.T) {
	model := PackageCleanupPolicyResourceModelV1{
		PackageCleanupPolicyResourceModelV0: PackageCleanupPolicyResourceModelV0{
			CronExpression: types.StringNull(),
		},
	}

	diags := model.fromAPIModel(context.Background(), PackageCleanupPolicyAPIModel{
		CronExpression: "",
		SearchCriteria: packageCleanupPolicyTestSearchCriteria(),
	})

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if !model.CronExpression.IsNull() {
		t.Fatalf("expected null cron_expression, got %q", model.CronExpression.ValueString())
	}
}

func TestPackageCleanupPolicyFromAPIModelKeepsConfiguredEmptyCronExpression(t *testing.T) {
	model := PackageCleanupPolicyResourceModelV1{
		PackageCleanupPolicyResourceModelV0: PackageCleanupPolicyResourceModelV0{
			CronExpression: types.StringValue(""),
		},
	}

	diags := model.fromAPIModel(context.Background(), PackageCleanupPolicyAPIModel{
		CronExpression: "",
		SearchCriteria: packageCleanupPolicyTestSearchCriteria(),
	})

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if model.CronExpression.IsNull() {
		t.Fatal("expected configured empty cron_expression to stay as an empty string")
	}

	if model.CronExpression.ValueString() != "" {
		t.Fatalf("expected empty cron_expression, got %q", model.CronExpression.ValueString())
	}
}

func TestPackageCleanupPolicyFromAPIModelSetsNonEmptyCronExpression(t *testing.T) {
	model := PackageCleanupPolicyResourceModelV1{
		PackageCleanupPolicyResourceModelV0: PackageCleanupPolicyResourceModelV0{
			CronExpression: types.StringNull(),
		},
	}

	diags := model.fromAPIModel(context.Background(), PackageCleanupPolicyAPIModel{
		CronExpression: "0 0 2 ? * MON-SAT *",
		SearchCriteria: packageCleanupPolicyTestSearchCriteria(),
	})

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if model.CronExpression.ValueString() != "0 0 2 ? * MON-SAT *" {
		t.Fatalf("expected non-empty cron_expression, got %q", model.CronExpression.ValueString())
	}
}

func TestPackageCleanupPolicyFromAPIModelLeavesFolderPathsNullWhenUnset(t *testing.T) {
	model := PackageCleanupPolicyResourceModelV1{}

	diags := model.fromAPIModel(context.Background(), PackageCleanupPolicyAPIModel{
		SearchCriteria: packageCleanupPolicyTestSearchCriteria(),
	})

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	attrs := model.SearchCriteria.Attributes()
	for _, key := range []string{"included_folder_paths", "excluded_folder_paths"} {
		if !attrs[key].IsNull() {
			t.Fatalf("expected %q to be null when not returned by the API, got %s", key, attrs[key])
		}
	}
}

func TestPackageCleanupPolicyFromAPIModelSetsFolderPaths(t *testing.T) {
	model := PackageCleanupPolicyResourceModelV1{}

	searchCriteria := packageCleanupPolicyTestSearchCriteria()
	searchCriteria.IncludedFolderPaths = &[]string{"*/staging/*"}
	searchCriteria.ExcludedFolderPaths = &[]string{"*/release/*"}

	diags := model.fromAPIModel(context.Background(), PackageCleanupPolicyAPIModel{
		SearchCriteria: searchCriteria,
	})

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	attrs := model.SearchCriteria.Attributes()
	expected := map[string]string{
		"included_folder_paths": "*/staging/*",
		"excluded_folder_paths": "*/release/*",
	}
	for key, want := range expected {
		set, ok := attrs[key].(types.Set)
		if !ok {
			t.Fatalf("expected %q to be a Set, got %T", key, attrs[key])
		}
		elements := set.Elements()
		if len(elements) != 1 {
			t.Fatalf("expected %q to have 1 element, got %d", key, len(elements))
		}
		got, ok := elements[0].(types.String)
		if !ok || got.ValueString() != want {
			t.Fatalf("expected %q to contain %q, got %s", key, want, elements[0])
		}
	}
}

func packageCleanupPolicyTestSearchCriteria() PackageCleanupPolicySearchCriteriaAPIModel {
	return PackageCleanupPolicySearchCriteriaAPIModel{
		PackageTypes: []string{"docker"},
		Repos:        []string{"example-repo"},
	}
}
