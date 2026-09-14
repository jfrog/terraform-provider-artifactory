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

package security

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestScopedTokenScopesValidatorAllowsSystemStorageInfo(t *testing.T) {
	validateScopedTokenScopes(t, "system:info/storage:r", false)
}

func TestScopedTokenScopesValidatorRejectsUnsupportedSystemStorageInfoAction(t *testing.T) {
	validateScopedTokenScopes(t, "system:info/storage:w", true)
}

func validateScopedTokenScopes(t *testing.T, scope string, expectError bool) {
	t.Helper()

	scopesAttribute, ok := schemaAttributesV1["scopes"].(schema.SetAttribute)
	if !ok {
		t.Fatal("expected scopes attribute to be a schema.SetAttribute")
	}

	scopesValue, diags := types.SetValue(types.StringType, []attr.Value{
		types.StringValue(scope),
	})
	if diags.HasError() {
		t.Fatalf("failed to build scopes value: %s", diags.Errors())
	}

	for _, setValidator := range scopesAttribute.Validators {
		var response validator.SetResponse
		setValidator.ValidateSet(context.Background(), validator.SetRequest{
			Path:        path.Root("scopes"),
			ConfigValue: scopesValue,
		}, &response)

		if expectError && response.Diagnostics.HasError() {
			return
		}

		if !expectError && response.Diagnostics.HasError() {
			t.Fatalf("expected %q to pass validation, got: %s", scope, response.Diagnostics.Errors())
		}
	}

	if expectError {
		t.Fatalf("expected %q to fail validation", scope)
	}
}
