package main

import (
	"github.com/atlassian/terraform-provider-artifactory/pkg/artifactory"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: artifactory.Provider,
	})
}
