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
