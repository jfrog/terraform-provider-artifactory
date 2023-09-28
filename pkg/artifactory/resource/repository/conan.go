package repository

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const ConanPackageType = "conan"

type ConanBaseParams struct {
	EnableConanSupport       bool `json:"enableConanSupport"`
	ForceConanAuthentication bool `json:"forceConanAuthentication"`
}

var ConanBaseSchema = map[string]*schema.Schema{
	"force_conan_authentication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Force basic authentication credentials in order to use this repository. Default value is 'false'.",
	},
}
