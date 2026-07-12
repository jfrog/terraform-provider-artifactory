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

package repository

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ConanResourceModel struct {
	ForceConanAuthentication types.Bool `tfsdk:"force_conan_authentication"`
}

type ConanAPIModel struct {
	EnableConanSupport       bool `json:"enableConanSupport"`
	ForceConanAuthentication bool `json:"forceConanAuthentication"`
}

var ConanAttributes = map[string]schema.Attribute{
	"force_conan_authentication": schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		MarkdownDescription: "Force basic authentication credentials in order to use this repository. Default value is 'false'.",
	},
}

// SDKv2
type ConanBaseParams struct {
	EnableConanSupport       bool `json:"enableConanSupport"`
	ForceConanAuthentication bool `json:"forceConanAuthentication"`
}

var ConanBaseSchemaSDKv2 = map[string]*sdkv2_schema.Schema{
	"force_conan_authentication": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Force basic authentication credentials in order to use this repository. Default value is 'false'.",
	},
}
