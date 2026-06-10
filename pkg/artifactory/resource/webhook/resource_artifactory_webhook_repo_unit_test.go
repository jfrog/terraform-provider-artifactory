// Copyright (c) JFrog Ltd. (2025)
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

package webhook

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFromRepoCriteriaAPIModelDefaultsMissingAnyFederatedToFalse(t *testing.T) {
	ctx := context.Background()
	baseCriteriaAttrs := map[string]attr.Value{
		"include_patterns": types.SetNull(types.StringType),
		"exclude_patterns": types.SetNull(types.StringType),
	}
	criteriaAPIModel := map[string]interface{}{
		"anyLocal":  false,
		"anyRemote": false,
		"repoKeys":  []interface{}{"my-local-repo"},
	}

	criteriaSet, diags := fromRepoCriteriaAPIMode(ctx, criteriaAPIModel, baseCriteriaAttrs)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	criteria := criteriaSet.Elements()[0].(types.Object)
	anyFederated := criteria.Attributes()["any_federated"].(types.Bool)
	if anyFederated.ValueBool() {
		t.Fatal("expected missing anyFederated API field to default to false")
	}
}
