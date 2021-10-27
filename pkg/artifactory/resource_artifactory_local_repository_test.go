package artifactory

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccLocalAlpineRepository(t *testing.T) {
	_, fqrn, name := mkNames("terraform-local-test-repo-basic", "artifactory_local_alpine_repository")
	kpId, kpFqrn, kpName := mkNames("some-keypair", "artifactory_keypair")
	localRepositoryBasic := executeTemplate("keypair",`
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "RSA"
			alias = "foo-alias{{ .kp_id }}"
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
		resource "artifactory_local_alpine_repository" "{{ .repo_name }}" {
			key 	     = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name 
			depends_on = [artifactory_keypair.{{ .kp_name }}]
		}
	`, map[string]interface{}{
		"kp_id" : kpId,
		"kp_name": kpName,
		"repo_name" : name,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		CheckDestroy: compositeCheckDestroy(
			verifyDeleted(fqrn, testCheckRepo),
			verifyDeleted(kpFqrn, verifyKeyPair),
		),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "alpine"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
				),
			},
		},
	})
}

func TestAccLocalRepository_basic(t *testing.T) {
	name := fmt.Sprintf("terraform-local-test-repo-basic%d", rand.Int())
	resourceName := fmt.Sprintf("artifactory_local_repository.%s", name)
	localRepositoryBasic := fmt.Sprintf(`
		resource "artifactory_local_repository" "%s" {
			key 	     = "%s"
			package_type = "docker"
		}
	`, name, name) // we use randomness so that, in the case of failure and dangle, the next test can run without collision
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName,testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", "docker"),
				),
			},
		},
	})
}

func mkTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-local-test-%d-full", rand.Int())
	resourceName := fmt.Sprintf("artifactory_local_repository.%s", name)
	const localRepositoryConfigFull = `
		resource "artifactory_local_repository" "%s" {
			key                             = "%s"
			package_type                    = "%s"
			description                     = "Test repo for %s"
			notes                           = "Test repo for %s"
			includes_pattern                = "**/*"
			excludes_pattern                = "**/*.tgz"
			repo_layout_ref                 = "npm-default"
			handle_releases                 = true
			handle_snapshots                = true
			max_unique_snapshots            = 25
			debian_trivial_layout           = false
			checksum_policy_type            = "client-checksums"
			max_unique_tags                 = 100
			snapshot_version_behavior       = "unique"
			suppress_pom_consistency_checks = true
			blacked_out                     = false
			property_sets                   = [ "artifactory" ]
			archive_browsing_enabled        = false
			calculate_yum_metadata          = false
			yum_root_depth                  = 0
			docker_api_version              = "V2"
		}
	`

	cfg := fmt.Sprintf(localRepositoryConfigFull, name, name, repoType, name, name)
	return t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName,testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "includes_pattern", "**/*"),
					resource.TestCheckResourceAttr(resourceName, "excludes_pattern", "**/*.tgz"),
					resource.TestCheckResourceAttr(resourceName, "repo_layout_ref", "npm-default"),
					resource.TestCheckResourceAttr(resourceName, "handle_releases", "true"),
					resource.TestCheckResourceAttr(resourceName, "handle_snapshots", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_unique_snapshots", "25"),
					resource.TestCheckResourceAttr(resourceName, "debian_trivial_layout", "false"),
					resource.TestCheckResourceAttr(resourceName, "checksum_policy_type", "client-checksums"),
					resource.TestCheckResourceAttr(resourceName, "max_unique_tags", "100"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_version_behavior", "unique"),
					resource.TestCheckResourceAttr(resourceName, "suppress_pom_consistency_checks", "true"),
					resource.TestCheckResourceAttr(resourceName, "blacked_out", "false"),
					resource.TestCheckResourceAttr(resourceName, "property_sets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "property_sets.0", "artifactory"),
					resource.TestCheckResourceAttr(resourceName, "archive_browsing_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "calculate_yum_metadata", "false"),
					resource.TestCheckResourceAttr(resourceName, "yum_root_depth", "0"),
					resource.TestCheckResourceAttr(resourceName, "docker_api_version", "V2"),
				),
			},
		},
	}
}

func TestAccAllRepoTypesLocal(t *testing.T) {

	for _, repo := range repoTypesSupported {
		t.Run(fmt.Sprintf("TestLocal%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
			resource.Test(mkTestCase(repo, t))
		})
	}
}
