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

package federated

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testMember1Url = "https://rt1.example.com/artifactory/test-repo"
	testMember2Url = "https://rt2.example.com/artifactory/test-repo"
)

func memberOnlySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{"member": SchemaGeneratorV4(true)["member"]}
}

func newMemberResourceData(t *testing.T, members []interface{}) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, memberOnlySchema(), map[string]interface{}{"member": members})
}

func memberStateByUrl(t *testing.T, d *schema.ResourceData) map[string]map[string]interface{} {
	t.Helper()
	result := map[string]map[string]interface{}{}
	for _, m := range d.Get("member").(*schema.Set).List() {
		member := m.(map[string]interface{})
		result[member["url"].(string)] = member
	}
	return result
}

func TestMemberJSON_ModeOmittedWhenEmpty(t *testing.T) {
	withoutMode, err := json.Marshal(Member{Url: testMember1Url, Enabled: true})
	require.NoError(t, err)
	assert.JSONEq(t, `{"url":"`+testMember1Url+`","enabled":true}`, string(withoutMode))

	withMode, err := json.Marshal(Member{Url: testMember2Url, Enabled: true, Mode: memberModeReceiver})
	require.NoError(t, err)
	assert.JSONEq(t, `{"url":"`+testMember2Url+`","enabled":true,"mode":"receiver"}`, string(withMode))
}

func TestUnpackMembers_Mode(t *testing.T) {
	d := newMemberResourceData(t, []interface{}{
		map[string]interface{}{"url": testMember1Url, "enabled": true},
		map[string]interface{}{"url": testMember2Url, "enabled": true, "mode": memberModeReceiver},
	})

	byUrl := map[string]Member{}
	for _, m := range unpackMembers(d) {
		byUrl[m.Url] = m
	}

	require.Len(t, byUrl, 2)
	assert.Empty(t, byUrl[testMember1Url].Mode)
	assert.Equal(t, memberModeReceiver, byUrl[testMember2Url].Mode)
}

func TestPackMembers_Mode(t *testing.T) {
	testCases := []struct {
		name         string
		apiMode      string
		expectedMode interface{}
	}{
		{name: "receiver is stored", apiMode: "receiver", expectedMode: "receiver"},
		{name: "empty is unset", apiMode: "", expectedMode: ""},
		{name: "bidirectional is unset", apiMode: "bidirectional", expectedMode: ""},
		{name: "transmitter is unset", apiMode: "transmitter", expectedMode: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d := newMemberResourceData(t, []interface{}{
				map[string]interface{}{"url": testMember1Url, "enabled": true},
			})

			err := PackMembers([]Member{{Url: testMember2Url, Enabled: true, Mode: tc.apiMode}}, d)
			require.NoError(t, err)

			member := memberStateByUrl(t, d)[testMember2Url]
			require.NotNil(t, member)
			assert.Equal(t, tc.expectedMode, member["mode"])
		})
	}
}

func TestPackUnpackMembers_RoundTrip(t *testing.T) {
	d := newMemberResourceData(t, []interface{}{
		map[string]interface{}{"url": testMember1Url, "enabled": true},
		map[string]interface{}{"url": testMember2Url, "enabled": false, "mode": memberModeReceiver},
	})

	members := unpackMembers(d)
	require.Len(t, members, 2)

	require.NoError(t, PackMembers(members, d))

	state := memberStateByUrl(t, d)
	assert.Equal(t, "", state[testMember1Url]["mode"])
	assert.Equal(t, memberModeReceiver, state[testMember2Url]["mode"])
	assert.Equal(t, false, state[testMember2Url]["enabled"])
}

func TestMemberSchema_ModeValidation(t *testing.T) {
	testCases := []struct {
		name      string
		mode      interface{}
		expectErr bool
	}{
		{name: "omitted", mode: nil, expectErr: false},
		{name: "receiver", mode: "receiver", expectErr: false},
		{name: "transmitter", mode: "transmitter", expectErr: true},
		{name: "bidirectional", mode: "bidirectional", expectErr: true},
		{name: "wrong case", mode: "Receiver", expectErr: true},
		{name: "empty string", mode: "", expectErr: true},
	}

	res := &schema.Resource{Schema: memberOnlySchema()}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			member := map[string]interface{}{"url": testMember2Url, "enabled": true}
			if tc.mode != nil {
				member["mode"] = tc.mode
			}

			diags := res.Validate(terraform.NewResourceConfigRaw(map[string]interface{}{
				"member": []interface{}{member},
			}))

			assert.Equal(t, tc.expectErr, diags.HasError(), "diagnostics: %v", diags)
		})
	}
}

func TestUpgradeMemberMode(t *testing.T) {
	rawState := map[string]any{
		"key": "test-repo",
		"member": []interface{}{
			map[string]interface{}{"url": testMember1Url, "enabled": true, "access_token": nil},
			map[string]interface{}{"url": testMember2Url, "enabled": true, "access_token": "token"},
		},
	}

	upgraded, err := upgradeMemberMode(context.Background(), rawState, nil)
	require.NoError(t, err)

	members := upgraded["member"].([]interface{})
	require.Len(t, members, 2)
	for _, m := range members {
		member := m.(map[string]interface{})
		value, present := member["mode"]
		assert.True(t, present, "mode key must be present after upgrade")
		assert.Nil(t, value)
	}
	assert.Equal(t, "token", members[1].(map[string]interface{})["access_token"])
}

func TestUpgradeMemberMode_NoMembers(t *testing.T) {
	rawState := map[string]any{"key": "test-repo"}

	upgraded, err := upgradeMemberMode(context.Background(), rawState, nil)
	require.NoError(t, err)
	assert.Equal(t, rawState, upgraded)
}

func TestResourceV4_HasNoMemberMode(t *testing.T) {
	memberElem := resourceV4().Schema["member"].Elem.(*schema.Resource)
	assert.NotContains(t, memberElem.Schema, "mode")
	assert.Contains(t, memberElem.Schema, "access_token")

	currentElem := SchemaGeneratorV4(true)["member"].Elem.(*schema.Resource)
	assert.Contains(t, currentElem.Schema, "mode")
}
