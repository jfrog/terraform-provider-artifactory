package local_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccDataSourceLocalHexRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-local-test-repo-basic", "data.artifactory_local_hex_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")
	localRepositoryBasic := util.ExecuteTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
{{ .private_key }}
EOF
			public_key = <<EOF
{{ .public_key }}
EOF
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key 	     = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			depends_on = [artifactory_keypair.{{ .kp_name }}]
		}

		data "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = artifactory_local_hex_repository.{{ .repo_name }}.key
		}
	`, map[string]interface{}{
		"kp_id":       kpId,
		"kp_name":     kpName,
		"repo_name":   name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "hex"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string { r, _ := repository.GetDefaultRepoLayoutRef("local", "hex"); return r }()),
				),
			},
		},
	})
}
