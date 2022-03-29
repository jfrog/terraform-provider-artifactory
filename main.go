package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/jfrog/terraform-provider-artifactory/v3/pkg/artifactory"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: artifactory.Provider,
	})
}
