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

package virtual_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/virtual"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualSchema_HideUnauthorizedResources(t *testing.T) {
	t.Parallel()

	_, ok := virtual.BaseSchemaV1["hide_unauthorized_resources"]
	assert.True(t, ok, "SDKv2 base schema must expose hide_unauthorized_resources")

	_, ok = virtual.VirtualAttributes["hide_unauthorized_resources"]
	assert.True(t, ok, "Framework virtual attributes must expose hide_unauthorized_resources")

	mavenResource := virtual.ResourceArtifactoryVirtualJavaRepository("maven")
	_, ok = mavenResource.Schema["hide_unauthorized_resources"]
	assert.True(t, ok, "virtual maven repository schema must expose hide_unauthorized_resources")
}

func TestRepositoryBaseParams_HideUnauthorizedResourcesJSON(t *testing.T) {
	t.Parallel()

	params := virtual.RepositoryBaseParams{
		Key:                       "maven-virt",
		Rclass:                    virtual.Rclass,
		PackageType:               "maven",
		HideUnauthorizedResources: true,
	}

	raw, err := json.Marshal(params)
	require.NoError(t, err)

	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, true, decoded["hideUnauthorizedResources"])

	var roundTrip virtual.RepositoryBaseParams
	require.NoError(t, json.Unmarshal(raw, &roundTrip))
	assert.True(t, roundTrip.HideUnauthorizedResources)
}

func TestVirtualResourceModel_HideUnauthorizedResourcesRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	model := virtual.VirtualResourceModel{
		BaseResourceModel: repository.BaseResourceModel{
			Key:                 types.StringValue("maven-virt"),
			ProjectKey:          types.StringValue(""),
			ProjectEnvironments: types.SetValueMust(types.StringType, []attr.Value{}),
			Description:         types.StringValue(""),
			Notes:               types.StringValue(""),
			IncludesPattern:     types.StringValue("**/*"),
			ExcludesPattern:     types.StringValue(""),
		},
		Repositories: types.ListValueMust(types.StringType, []attr.Value{}),
		ArtifactoryRequestsCanRetrieveRemoteArtifacts: types.BoolValue(false),
		DefaultDeploymentRepo:                         types.StringValue(""),
		RepoLayoutRef:                                 types.StringValue("maven-2-default"),
		HideUnauthorizedResources:                     types.BoolValue(true),
	}

	apiModel, diags := model.ToAPIModel(ctx, "maven")
	require.False(t, diags.HasError(), diags.Errors())

	virtualAPI, ok := apiModel.(virtual.VirtualAPIModel)
	require.True(t, ok)
	assert.True(t, virtualAPI.HideUnauthorizedResources)

	raw, err := json.Marshal(virtualAPI)
	require.NoError(t, err)
	var decoded map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, true, decoded["hideUnauthorizedResources"])

	var fromAPI virtual.VirtualResourceModel
	fromAPI.Repositories = types.ListNull(types.StringType)
	diags = fromAPI.FromAPIModel(ctx, virtualAPI)
	require.False(t, diags.HasError(), diags.Errors())
	assert.True(t, fromAPI.HideUnauthorizedResources.ValueBool())
}
