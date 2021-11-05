package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/jfrog/terraform-provider-xray/pkg/xray"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: xray.Provider,
	})
}
