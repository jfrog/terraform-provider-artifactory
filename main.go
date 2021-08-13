package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"

	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: artifactory.Provider,
	})
}
