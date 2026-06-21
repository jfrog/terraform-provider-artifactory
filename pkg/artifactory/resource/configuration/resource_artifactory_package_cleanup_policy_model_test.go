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

func packageCleanupPolicyTestSearchCriteria() PackageCleanupPolicySearchCriteriaAPIModel {
	return PackageCleanupPolicySearchCriteriaAPIModel{
		PackageTypes: []string{"docker"},
		Repos:        []string{"example-repo"},
	}
}
