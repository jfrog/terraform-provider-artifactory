package remote_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
)

func TestAccRemoteNvidiaNimRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-nvidia-nim-remote", "artifactory_remote_nvidia_nim_repository")

	const template = `
resource "artifactory_remote_nvidia_nim_repository" "{{ .name }}" {
  key = "{{ .name }}"
  url = "https://api.ngc.nvidia.com"
}
`
	config, err := testutil.ExecuteTemplate("TestAccRemoteNvidiaNimRepository_basic", template, map[string]string{"name": name})
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "nvidia-nim"),
					resource.TestCheckResourceAttr(fqrn, "url", "https://api.ngc.nvidia.com"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}