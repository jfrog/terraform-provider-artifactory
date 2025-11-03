package virtual_test

import (
	"bytes"
	"os"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
)

func TestAccVirtualHexRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("hex-repo", "artifactory_virtual_hex_repository")
	_, _, repoName := testutil.MkNames("local-repo", "artifactory_local_hex_repository")
	_, _, kpName := testutil.MkNames("some-keypair", "artifactory_keypair")

	tmpl := template.Must(template.New("test").Parse(`
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
		}

		resource "artifactory_local_hex_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			
			depends_on = [artifactory_keypair.{{ .kp_name }}]
		}

		resource "artifactory_virtual_hex_repository" "{{ .name }}" {
			key          = "{{ .name }}"
			repositories = [artifactory_local_hex_repository.{{ .repo_name }}.key]
			hex_primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			
			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_local_hex_repository.{{ .repo_name }}
			]
		}
	`))

	data := map[string]interface{}{
		"kp_name":     kpName,
		"kp_id":       kpName,
		"repo_name":   repoName,
		"name":        name,
		"private_key": os.Getenv("JFROG_TEST_RSA_PRIVATE_KEY"),
		"public_key":  os.Getenv("JFROG_TEST_RSA_PUBLIC_KEY"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	config := buf.String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6MuxProviderFactories,
		CheckDestroy:             acctest.VerifyDeleted(t, fqrn, "key", acctest.CheckRepo),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "hex_primary_keypair_ref", kpName),
				),
			},
		},
	})
}
