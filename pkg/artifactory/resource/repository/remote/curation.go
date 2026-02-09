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

package remote

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CurationResourceModel struct {
	Curated     types.Bool `tfsdk:"curated"`
	PassThrough types.Bool `tfsdk:"pass_through"`
}

type CurationAPIModel struct {
	Curated     bool `json:"curated"`
	PassThrough bool `json:"passThrough"`
}

var CurationAttributes = map[string]schema.Attribute{
	"curated": schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		MarkdownDescription: "Enable repository to be protected by the Curation service.",
	},
	"pass_through": schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		MarkdownDescription: "Enable Pass-through for Curation Audit. When enabled, allows artifacts to pass through the Curation audit process.",
	},
}

// SDKv2
type RepositoryCurationParams struct {
	Curated     bool `json:"curated"`
	PassThrough bool `json:"passThrough"`
}

var CurationRemoteRepoSchema = map[string]*sdkv2_schema.Schema{
	"curated": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable repository to be protected by the Curation service.",
	},
	"pass_through": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable Pass-through for Curation Audit. When enabled, allows artifacts to pass through the Curation audit process.",
	},
}
