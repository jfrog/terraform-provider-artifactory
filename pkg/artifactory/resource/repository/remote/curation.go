package remote

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkv2_schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type CurationResourceModel struct {
	Curated types.Bool `tfsdk:"curated"`
}

type CurationAPIModel struct {
	Curated bool `json:"curated"`
}

var CurationAttributes = map[string]schema.Attribute{
	"curated": schema.BoolAttribute{
		Optional:            true,
		Computed:            true,
		Default:             booldefault.StaticBool(false),
		MarkdownDescription: "Enable repository to be protected by the Curation service.",
	},
}

// SDKv2
type RepositoryCurationParams struct {
	Curated bool `json:"curated"`
}

var CurationRemoteRepoSchema = map[string]*sdkv2_schema.Schema{
	"curated": {
		Type:        sdkv2_schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable repository to be protected by the Curation service.",
	},
}
