package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/provider"
)

func TestMuxServer(t *testing.T) {
	const providerConfig = `
		provider "artifactory" {
			url        		= "%s"
			access_token    = "%s"
			check_license   = true
		}
	`
	url := os.Getenv("JFROG_URL")
	token := os.Getenv("JFROG_ACCESS_TOKEN")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: map[string]func() (tfprotov5.ProviderServer, error){
			"artifactory": func() (tfprotov5.ProviderServer, error) {
				ctx := context.Background()
				providers := []func() tfprotov5.ProviderServer{
					providerserver.NewProtocol5(provider.Framework()()), // terraform-plugin-framework provider
					provider.SdkV2().GRPCProvider,                       // terraform-plugin-sdk provider
				}

				muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)

				if err != nil {
					return nil, err
				}

				return muxServer.ProviderServer(), nil
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(providerConfig, url, token),
			},
		},
	})
}
