package provider_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/provider"
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
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"artifactory": func() (tfprotov6.ProviderServer, error) {
				ctx := context.Background()

				upgradedSdkServer, err := tf5to6server.UpgradeServer(
					ctx,
					provider.SdkV2().GRPCProvider, // terraform-plugin-sdk provider
				)
				if err != nil {
					return nil, err
				}

				providers := []func() tfprotov6.ProviderServer{
					providerserver.NewProtocol6(provider.Framework()()), // terraform-plugin-framework provider
					func() tfprotov6.ProviderServer {
						return upgradedSdkServer
					},
				}

				muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

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
