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

package remote

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRemoteHuggingFaceMLSchema_EnableTokenAuthenticationDefaultsTrue(t *testing.T) {
	r := NewHuggingFaceMLRemoteRepositoryResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["enable_token_authentication"].(schema.BoolAttribute)
	if !ok {
		t.Fatal("enable_token_authentication schema attribute missing")
	}
	if attr.Default == nil {
		t.Fatal("enable_token_authentication has no default")
	}

	var dresp defaults.BoolResponse
	attr.Default.DefaultBool(context.Background(), defaults.BoolRequest{}, &dresp)
	if !dresp.PlanValue.Equal(types.BoolValue(true)) {
		t.Fatalf("enable_token_authentication default = %v, want true", dresp.PlanValue)
	}
}

func TestRemoteAIEditorExtensionsModels_NoDuplicateEnableTokenAuthenticationTags(t *testing.T) {
	assertNoDuplicateStructTags(t, RemoteAIEditorExtensionsResourceModel{}, "tfsdk")
	assertNoDuplicateStructTags(t, RemoteAIEditorExtensionsAPIModel{}, "json")
}

func assertNoDuplicateStructTags(t *testing.T, v any, tagKey string) {
	t.Helper()

	seen := map[string]string{}
	var walk func(reflect.Type, string)
	walk = func(typ reflect.Type, prefix string) {
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		if typ.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			path := prefix + field.Name
			if field.Anonymous && (field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct)) {
				walk(field.Type, path+".")
			}
			tag := field.Tag.Get(tagKey)
			if tag == "" || tag == "-" {
				continue
			}
			if comma := strings.IndexByte(tag, ','); comma >= 0 {
				tag = tag[:comma]
			}
			if prev, ok := seen[tag]; ok {
				t.Errorf("duplicate %s %q on %s and %s", tagKey, tag, prev, path)
			}
			seen[tag] = path
		}
	}
	walk(reflect.TypeOf(v), "")
}
