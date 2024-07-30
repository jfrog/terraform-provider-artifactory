package lifecycle_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v11/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccReleaseBundleV2Promotion_full(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2-promotion", "artifactory_release_bundle_v2_promotion")
	_, _, releaseBundleName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	_, _, err := uploadTestFile(t, repoName)
	if err != nil {
		t.Fatalf("failed to upload file: %s", err)
	}

	keyPairName := fmt.Sprintf("test-keypair-%d", testutil.RandomInt())

	const template = `
	resource "artifactory_keypair" "{{ .keypair_name }}" {
		pair_name = "{{ .keypair_name }}"
		pair_type = "RSA"
		alias = "test-alias-{{ .keypair_name }}"
		private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA2ymVc24BoaZb0ElXoI3X4zUKJGZEetR6F4yT1tJhkPDg7UTm
iNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbqFUaXPgud8VabfHl0imXvN746zmpj
YEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEppj5yN0tVWDnqjOJjR7EpxMSdP3TSd
6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQ
FmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDMX8KGs7e9ZgjANkT5PnipLOaeLJU4
H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJDQIDAQABAoIBAFb7wyhEIfuhhlE9
uryrb2LrGzJlMIq7qBWOouKhLz4SjIM/VGw+c76VkjZGoSU+LeLj+D0W1ie0u2Cw
gJM8aW22TbK/c5lksWOO5PVFDdPG+ZoRWY3MLGlhlL5E4UhMPgJyy/eeiRjZ3CZM
pja+UfVAwn1KVNR8UinVZYPt680AvEd1FGxoNLxemIPNug46nNqp6Al86Bn+BnkN
GXpwyooXVSfo4k+pnFBFIXUdA1dUEXQSVb1AxlTo6g/Ok/+8Gfq8idCdu+5fcZI2
1eLeC+FAa92rr1SFX/UWeB4cMyuAqwuxbFFIplT22SaUSsNuOUSH4E00nbP/AuCb
1BqrLmECgYEA7tQKfyF9aiXTsOMdOnSAa5OyEaCfsFtcmLd4ykVrwN8O36NoX005
VbGuqo87fwIXQIKHU+kOEs/TmaQ8bNcbCD/SfWGTtOnHG4qfIsepJuoMwbQHRhGF
JeoXh5yEUKg5pcDBY8PENEtEVKmFuL4bPOdn+9FNLGsjftvXpmGWbGUCgYEA6uuQ
7kzO6WD88OsxdJzlJM11hg2SaSBCh3+5tnOhF1ULOUt4tdYXzh3QI6BPX7tkArYf
XteVfWoWqn6T7LtCjFm350BqVpPhqfLKnt6fYf1yotsj/cyZXlXquRbxbgakB0n0
4PrsZaube0TPPVeirzNyOVHyFc+iW+F+IUYD+4kCgYEApDEjBkP/9PoMj4+UiJuP
rmXcBkJnhtdI0bVRVb5kVjUEBLxTBTISONfvPVM7lBXb5n3Wi9mt00EOOJKw+CLq
csFt9MUgxz/xov2qaj7aC+bc3k7msUVaRLardpAkZ09AUrQyQGRWf50/XPUu+dO4
5iYxVu6OH/uIa664k6qDwAECgYEAslS8oomgEL3VhbWkx1dLA5MMggTPfgpFNsMY
4Y4JXcLrUEUgjzjEvW0YUdMiLhP8qapDSiXxj1D3f9myxWSp8g0xc9UMZEjCZ9at
RcjNyP8zBLnCKqokSt6B3puyDsnvvrC/ugIBbnTFBOCJSZG7J7CwJx8z3KbQI1ub
+fpCj7ECgYAd69soLEybUGMjsdI+OijIGoUTUoZGXJm+0VpBt4QJCe7AMnYPfYzA
JnEmN4D7HLTKUBklQnb/FhP/RuiT2bSAd1l+PNeuU7gYROCBBonzxXQ1wEbNrSYA
iyoc9g/kvV8HTW8361xEhu7wmuSEEx1gQ/7sdhTkgrTncc8hxVRxuA==
-----END RSA PRIVATE KEY-----
EOF
		public_key = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2ymVc24BoaZb0ElXoI3X
4zUKJGZEetR6F4yT1tJhkPDg7UTmiNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbq
FUaXPgud8VabfHl0imXvN746zmpjYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEpp
j5yN0tVWDnqjOJjR7EpxMSdP3TSd6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof
+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDM
X8KGs7e9ZgjANkT5PnipLOaeLJU4H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJ
DQIDAQAB
-----END PUBLIC KEY-----
EOF
	}
	
	resource "artifactory_release_bundle_v2" "{{ .release_bundle_name }}" {
		name = "{{ .release_bundle_name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}

	resource "artifactory_release_bundle_v2_promotion" "{{ .name }}" {
		name = artifactory_release_bundle_v2.{{ .release_bundle_name }}.name
		version = artifactory_release_bundle_v2.{{ .release_bundle_name }}.version
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		environment = "DEV"
		included_repository_keys = ["{{ .repo_name }}"]
	}`

	testData := map[string]string{
		"name":                resourceName,
		"keypair_name":        keyPairName,
		"repo_name":           repoName,
		"release_bundle_name": releaseBundleName,
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["release_bundle_name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "environment", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.0", testData["repo_name"]),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2Promotion_full_with_project(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2-promotion", "artifactory_release_bundle_v2_promotion")
	_, _, releaseBundleName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	_, _, projectName := testutil.MkNames("test-project-", "project")
	projectKey := fmt.Sprintf("test%d", testutil.RandomInt())

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	_, _, err := uploadTestFile(t, repoName)
	if err != nil {
		t.Fatalf("failed to upload file: %s", err)
	}

	keyPairName := fmt.Sprintf("test-keypair-%d", testutil.RandomInt())

	const template = `
	resource "project" "{{ .project_name }}" {
		key          = "{{ .project_key }}"
		display_name = "{{ .project_key }}"
		admin_privileges {
			manage_members   = true
			manage_resources = true
			index_resources  = true
		}
	}

	resource "project_repository" "{{ .project_repo_name }}" {
		project_key = project.{{ .project_name }}.key
		key         = "{{ .repo_name }}"
	}

	resource "artifactory_keypair" "{{ .keypair_name }}" {
		pair_name = "{{ .keypair_name }}"
		pair_type = "RSA"
		alias = "test-alias-{{ .keypair_name }}"
		private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA2ymVc24BoaZb0ElXoI3X4zUKJGZEetR6F4yT1tJhkPDg7UTm
iNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbqFUaXPgud8VabfHl0imXvN746zmpj
YEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEppj5yN0tVWDnqjOJjR7EpxMSdP3TSd
6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQ
FmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDMX8KGs7e9ZgjANkT5PnipLOaeLJU4
H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJDQIDAQABAoIBAFb7wyhEIfuhhlE9
uryrb2LrGzJlMIq7qBWOouKhLz4SjIM/VGw+c76VkjZGoSU+LeLj+D0W1ie0u2Cw
gJM8aW22TbK/c5lksWOO5PVFDdPG+ZoRWY3MLGlhlL5E4UhMPgJyy/eeiRjZ3CZM
pja+UfVAwn1KVNR8UinVZYPt680AvEd1FGxoNLxemIPNug46nNqp6Al86Bn+BnkN
GXpwyooXVSfo4k+pnFBFIXUdA1dUEXQSVb1AxlTo6g/Ok/+8Gfq8idCdu+5fcZI2
1eLeC+FAa92rr1SFX/UWeB4cMyuAqwuxbFFIplT22SaUSsNuOUSH4E00nbP/AuCb
1BqrLmECgYEA7tQKfyF9aiXTsOMdOnSAa5OyEaCfsFtcmLd4ykVrwN8O36NoX005
VbGuqo87fwIXQIKHU+kOEs/TmaQ8bNcbCD/SfWGTtOnHG4qfIsepJuoMwbQHRhGF
JeoXh5yEUKg5pcDBY8PENEtEVKmFuL4bPOdn+9FNLGsjftvXpmGWbGUCgYEA6uuQ
7kzO6WD88OsxdJzlJM11hg2SaSBCh3+5tnOhF1ULOUt4tdYXzh3QI6BPX7tkArYf
XteVfWoWqn6T7LtCjFm350BqVpPhqfLKnt6fYf1yotsj/cyZXlXquRbxbgakB0n0
4PrsZaube0TPPVeirzNyOVHyFc+iW+F+IUYD+4kCgYEApDEjBkP/9PoMj4+UiJuP
rmXcBkJnhtdI0bVRVb5kVjUEBLxTBTISONfvPVM7lBXb5n3Wi9mt00EOOJKw+CLq
csFt9MUgxz/xov2qaj7aC+bc3k7msUVaRLardpAkZ09AUrQyQGRWf50/XPUu+dO4
5iYxVu6OH/uIa664k6qDwAECgYEAslS8oomgEL3VhbWkx1dLA5MMggTPfgpFNsMY
4Y4JXcLrUEUgjzjEvW0YUdMiLhP8qapDSiXxj1D3f9myxWSp8g0xc9UMZEjCZ9at
RcjNyP8zBLnCKqokSt6B3puyDsnvvrC/ugIBbnTFBOCJSZG7J7CwJx8z3KbQI1ub
+fpCj7ECgYAd69soLEybUGMjsdI+OijIGoUTUoZGXJm+0VpBt4QJCe7AMnYPfYzA
JnEmN4D7HLTKUBklQnb/FhP/RuiT2bSAd1l+PNeuU7gYROCBBonzxXQ1wEbNrSYA
iyoc9g/kvV8HTW8361xEhu7wmuSEEx1gQ/7sdhTkgrTncc8hxVRxuA==
-----END RSA PRIVATE KEY-----
EOF
		public_key = <<EOF
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2ymVc24BoaZb0ElXoI3X
4zUKJGZEetR6F4yT1tJhkPDg7UTmiNoFB5TZJvP6LBrrSwszkpZbxaVOkBrwrGbq
FUaXPgud8VabfHl0imXvN746zmpjYEMGqJzm+Gh6aBWOlnPdLuHhds/kcanFAEpp
j5yN0tVWDnqjOJjR7EpxMSdP3TSd6tNAY73LGNLNJQc6tSxh8nMIb4HhSWQSgfof
+FwcLGvs+mmyBq8Jz5Zy4BSCk1fQFmCnSGyzpyBD0vMd6eLHk2l0tm56DrlonbDM
X8KGs7e9ZgjANkT5PnipLOaeLJU4H+OWyBZUAT4hl/iRVvLwV81x7g0/O38kmPYJ
DQIDAQAB
-----END PUBLIC KEY-----
EOF
	}
	
	resource "artifactory_release_bundle_v2" "{{ .release_bundle_name }}" {
		name = "{{ .release_bundle_name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		project_key = project.{{ .project_name }}.key
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}

	resource "artifactory_release_bundle_v2_promotion" "{{ .name }}" {
		name = artifactory_release_bundle_v2.{{ .release_bundle_name }}.name
		version = artifactory_release_bundle_v2.{{ .release_bundle_name }}.version
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		project_key = project.{{ .project_name }}.key
		environment = "DEV"
		included_repository_keys = ["{{ .repo_name }}"]
	}`

	testData := map[string]string{
		"name":                resourceName,
		"keypair_name":        keyPairName,
		"repo_name":           repoName,
		"release_bundle_name": releaseBundleName,
		"project_name":        projectName,
		"project_key":         projectKey,
		"project_repo_name":   fmt.Sprintf("%s-%s", projectKey, repoName),
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"project": {
				Source: "jfrog/project",
			},
		},
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["release_bundle_name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "project_key", testData["project_key"]),
					resource.TestCheckResourceAttr(fqrn, "environment", "DEV"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "included_repository_keys.0", testData["repo_name"]),
				),
			},
		},
	})
}
