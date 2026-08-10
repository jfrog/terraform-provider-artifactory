package local_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
)

func TestAccLocalNvidiaNimRepository_basic(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-nvidia-nim-local", "artifactory_local_nvidia_nim_repository")

	const template = `
resource "artifactory_local_nvidia_nim_repository" "{{ .name }}" {
  key = "{{ .name }}"
}
`
	config, err := testutil.ExecuteTemplate("TestAccLocalNvidiaNimRepository_basic", template, map[string]string{"name": name})
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