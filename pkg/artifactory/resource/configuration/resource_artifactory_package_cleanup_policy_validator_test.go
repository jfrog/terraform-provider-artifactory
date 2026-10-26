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
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var cleanupSearchCriteriaValidatorTestAttrTypes = map[string]attr.Type{
	"created_before_in_days":           types.Int64Type,
	"last_downloaded_before_in_days":   types.Int64Type,
	"created_before_in_months":         types.Int64Type,
	"last_downloaded_before_in_months": types.Int64Type,
	"keep_last_n_versions":             types.Int64Type,
	"included_properties":              types.MapType{ElemType: types.ListType{ElemType: types.StringType}},
}

func cleanupSearchCriteriaValidatorTestObject(t *testing.T, ints map[string]int64, withProperties bool) types.Object {
	t.Helper()

	values := map[string]attr.Value{}
	for name, attrType := range cleanupSearchCriteriaValidatorTestAttrTypes {
		if attrType == types.Int64Type {
			values[name] = types.Int64Null()
		}
	}
	for name, value := range ints {
		values[name] = types.Int64Value(value)
	}

	properties := types.MapNull(types.ListType{ElemType: types.StringType})
	if withProperties {
		properties = types.MapValueMust(
			types.ListType{ElemType: types.StringType},
			map[string]attr.Value{
				"release.status": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("prod")}),
			},
		)
	}
	values["included_properties"] = properties

	return types.ObjectValueMust(cleanupSearchCriteriaValidatorTestAttrTypes, values)
}

func TestCleanupSearchCriteriaValidator(t *testing.T) {
	testCases := []struct {
		name           string
		ints           map[string]int64
		withProperties bool
		errorContains  string
	}{
		{
			name: "version-based alone",
			ints: map[string]int64{"keep_last_n_versions": 7},
		},
		{
			name: "time-based alone",
			ints: map[string]int64{"created_before_in_days": 730},
		},
		{
			name:           "time-based and properties-based",
			ints:           map[string]int64{"created_before_in_days": 30},
			withProperties: true,
		},
		{
			name: "version-based and time-based (days)",
			ints: map[string]int64{"created_before_in_days": 730, "keep_last_n_versions": 7},
		},
		{
			name: "version-based and time-based (months)",
			ints: map[string]int64{"created_before_in_months": 24, "keep_last_n_versions": 7},
		},
		{
			name:           "version-based and properties-based",
			ints:           map[string]int64{"keep_last_n_versions": 7},
			withProperties: true,
		},
		{
			name:           "version-based, time-based and properties-based",
			ints:           map[string]int64{"last_downloaded_before_in_days": 730, "keep_last_n_versions": 7},
			withProperties: true,
		},
		{
			name:          "no condition",
			errorContains: "A policy must specify at least one condition",
		},
		{
			name:          "days and months together",
			ints:          map[string]int64{"created_before_in_days": 30, "created_before_in_months": 1},
			errorContains: "Cannot use both days-based conditions",
		},
		{
			name:          "zero keep_last_n_versions",
			ints:          map[string]int64{"keep_last_n_versions": 0},
			errorContains: "must have a value greater than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.ObjectRequest{
				Path:        path.Root("search_criteria"),
				ConfigValue: cleanupSearchCriteriaValidatorTestObject(t, tc.ints, tc.withProperties),
			}
			resp := &validator.ObjectResponse{}

			cleanupSearchCriteriaValidator{}.ValidateObject(context.Background(), req, resp)

			if tc.errorContains == "" {
				if resp.Diagnostics.HasError() {
					t.Fatalf("unexpected diagnostics: %s", resp.Diagnostics.Errors())
				}
				return
			}

			if !resp.Diagnostics.HasError() {
				t.Fatalf("expected an error containing %q, got none", tc.errorContains)
			}
			if detail := resp.Diagnostics.Errors()[0].Detail(); !strings.Contains(detail, tc.errorContains) {
				t.Fatalf("expected an error containing %q, got %q", tc.errorContains, detail)
			}
		})
	}
}
