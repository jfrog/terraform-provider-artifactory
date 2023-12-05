package remote

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

type RepositoryCurationParams struct {
	Curated bool `json:"curated"`
}

var CurationRemoteRepoSchema = map[string]*schema.Schema{
	"curated": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "Enable repository to be protected by the Curation service.",
	},
}
