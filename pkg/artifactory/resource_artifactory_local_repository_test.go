package artifactory

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

//TODO: failing test - to fix
func TestAccLocalAlpineRepository(t *testing.T) {
	_, fqrn, name := mkNames("terraform-local-test-repo-basic", "artifactory_local_alpine_repository")
	kpId, kpFqrn, kpName := mkNames("some-keypair", "artifactory_keypair")
	localRepositoryBasic := executeTemplate("keypair", `
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
		"kp_id":     kpId,
		"kp_name":   kpName,
		"repo_name": name,
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

//TODO: failing test - to fix
func TestAccLocalDebianRepository(t *testing.T) {
	_, fqrn, name := mkNames("local-debian-repo", "artifactory_local_debian_repository")
	kpId, kpFqrn, kpName := mkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := mkNames("some-keypair2", "artifactory_keypair")
	localRepositoryBasic := executeTemplate("keypair", `
		resource "artifactory_keypair" "{{ .kp_name }}" {
			pair_name  = "{{ .kp_name }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id }}"
			private_key = <<EOF
		-----BEGIN PGP PRIVATE KEY BLOCK-----

		lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
		jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
		1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
		aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
		w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
		CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
		AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
		nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
		EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
		MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
		GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
		JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
		x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
		=qCzy
		-----END PGP PRIVATE KEY BLOCK-----
		EOF
			public_key = <<EOF
		-----BEGIN PGP PUBLIC KEY BLOCK-----

		mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
		jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
		PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
		CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
		frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
		Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
		6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
		NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
		hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
		AQ==
		=2kMe
		-----END PGP PUBLIC KEY BLOCK-----
		EOF
		}
		resource "artifactory_keypair" "{{ .kp_name2 }}" {
			pair_name  = "{{ .kp_name2 }}"
			pair_type = "GPG"
			alias = "foo-alias{{ .kp_id2 }}"
			private_key = <<EOF
		-----BEGIN PGP PRIVATE KEY BLOCK-----

		lIYEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
		jn7zib/+BwMCFjb4odY28+n0NWj7KZ53BkA0qzzqT9IpIfsW/tLNPTxYEFrDVbcF
		1CuiAgAhyUfBEr9HQaMJBLfIIvo/B3nlWvwWHkiQFuWpsnJ2pj8F8LQqQ2hyaXN0
		aWFuIEJvbmdpb3JubyA8Y2hyaXN0aWFuYkBqZnJvZy5jb20+iJoEExYKAEIWIQSS
		w8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbAwUJA8JnAAULCQgHAgMiAgEGFQoJ
		CAsCBBYCAwECHgcCF4AACgkQwL80hJIR2yRQDgD/X1t/hW9+uXdSY59FOClhQw/t
		AzTYjDW+KLKadYJ3RAIBALD53rj7EnrXsSqv9Vqj3mJ7O38eXu50P57tD8ErpHMD
		nIsEYYU7tRIKKwYBBAGXVQEFAQEHQCfT+jXHVkslGAJqVafoeWO8Nwz/oPPzNDJb
		EOASsMRcAwEIB/4HAwK+Wi8OaidLuvQ6yknLUspoRL8KJlQu0JkfLxj6Wl6GrRtf
		MdUBxaGUQX5UzMIqyYstgHKz2kBYvrJijWdOkkRuL82FySSh4yi/97FBikOBiHgE
		GBYKACAWIQSSw8jt+9pdVC3Gts7AvzSEkhHbJAUCYYU7tQIbDAAKCRDAvzSEkhHb
		JNR/AQCQjGWljmP8pYj6ohP8bOwVB4VE5qxjdfWQvBCUA0LFwgEAxLGVeT88pw3+
		x7Cwd7SsuxlIOOCIJssFnUhA9Qsq2wE=
		=qCzy
		-----END PGP PRIVATE KEY BLOCK-----
		EOF
			public_key = <<EOF
		-----BEGIN PGP PUBLIC KEY BLOCK-----

		mDMEYYU7tRYJKwYBBAHaRw8BAQdAZ8vVdEyrWGssb7cdreG5GDGv6taHX/vWQdDG
		jn7zib+0KkNocmlzdGlhbiBCb25naW9ybm8gPGNocmlzdGlhbmJAamZyb2cuY29t
		PoiaBBMWCgBCFiEEksPI7fvaXVQtxrbOwL80hJIR2yQFAmGFO7UCGwMFCQPCZwAF
		CwkIBwIDIgIBBhUKCQgLAgQWAgMBAh4HAheAAAoJEMC/NISSEdskUA4A/19bf4Vv
		frl3UmOfRTgpYUMP7QM02Iw1viiymnWCd0QCAQCw+d64+xJ617Eqr/Vao95iezt/
		Hl7udD+e7Q/BK6RzA7g4BGGFO7USCisGAQQBl1UBBQEBB0An0/o1x1ZLJRgCalWn
		6HljvDcM/6Dz8zQyWxDgErDEXAMBCAeIeAQYFgoAIBYhBJLDyO372l1ULca2zsC/
		NISSEdskBQJhhTu1AhsMAAoJEMC/NISSEdsk1H8BAJCMZaWOY/yliPqiE/xs7BUH
		hUTmrGN19ZC8EJQDQsXCAQDEsZV5PzynDf7HsLB3tKy7GUg44IgmywWdSED1Cyrb
		AQ==
		=2kMe
		-----END PGP PUBLIC KEY BLOCK-----
		EOF
		}
		resource "artifactory_local_debian_repository" "{{ .repo_name }}" {
			key 	     = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = ["bz2","lzma","xz"]
			trivial_layout = true
			depends_on = [artifactory_keypair.{{ .kp_name }}]
		}
	`, map[string]interface{}{
		"kp_id":     kpId,
		"kp_name":   kpName,
		"kp_id2":    kpId2,
		"kp_name2":  kpName2,
		"repo_name": name,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		CheckDestroy: compositeCheckDestroy(
			verifyDeleted(fqrn, testCheckRepo),
			verifyDeleted(kpFqrn, verifyKeyPair),
			verifyDeleted(kpFqrn2, verifyKeyPair),
		),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "debian"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "true"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.0", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.1", "lzma"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.2", "xz"),
				),
			},
		},
	})
}

func TestAccLegacyLocalRepository_basic(t *testing.T) {
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
		CheckDestroy:      verifyDeleted(resourceName, testCheckRepo),
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

func TestAccLocalDockerV1Repository(t *testing.T) {

	_, fqrn, name := mkNames("dockerv1-local", "artifactory_local_docker_v1_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := executeTemplate("TestAccLocalDockerv2Repository", `
		resource "artifactory_local_docker_v1_repository" "{{ .name }}" {
			key 	     = "{{ .name }}"
		}
	`, params)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", "false"),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", "1"),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", "0"),
				),
			},
		},
	})
}
func TestAccLocalDockerV2Repository(t *testing.T) {

	_, fqrn, name := mkNames("dockerv2-local", "artifactory_local_docker_v2_repository")
	params := map[string]interface{}{
		"block":     randBool(),
		"retention": randSelect(1, 5, 10),
		"max_tags":  randSelect(0, 5, 10),
		"name":      name,
	}
	localRepositoryBasic := executeTemplate("TestAccLocalDockerV2Repository", `
		resource "artifactory_local_docker_v2_repository" "{{ .name }}" {
			key 	     = "{{ .name }}"
			tag_retention = {{ .retention }}
			max_unique_tags = {{ .max_tags }}
			block_pushing_schema1 = {{ .block }}
		}
	`, params)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "block_pushing_schema1", fmt.Sprintf("%t", params["block"])),
					resource.TestCheckResourceAttr(fqrn, "tag_retention", fmt.Sprintf("%d", params["retention"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_tags", fmt.Sprintf("%d", params["max_tags"])),
				),
			},
		},
	})
}
func TestAccLocalNugetRepository(t *testing.T) {

	_, fqrn, name := mkNames("nuget-local", "artifactory_local_nuget_repository")
	params := map[string]interface{}{
		"force_nuget_authentication": randBool(),
		"max_unique_snapshots":       randSelect(0, 5, 10),
		"name":                       name,
	}
	localRepositoryBasic := executeTemplate("TestAccLocalNugetRepository", `
		resource "artifactory_local_nuget_repository" "{{ .name }}" {
		  key                 = "{{ .name }}"
		  max_unique_snapshots = {{ .max_unique_snapshots }}
		  force_nuget_authentication = {{ .force_nuget_authentication }}
		}
	`, params)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", params["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", fmt.Sprintf("%t", params["force_nuget_authentication"])),
				),
			},
		},
	})
}

var commonJavaParams = map[string]interface{}{
	"name":                            "",
	"checksum_policy_type":            "client-checksums",
	"snapshot_version_behavior":       "unique",
	"max_unique_snapshots":            randSelect(0, 5, 10),
	"handle_releases":                 true,
	"handle_snapshots":                true,
	"suppress_pom_consistency_checks": false,
}

const localJavaRepositoryBasic = `
		resource "{{ .resource_name }}" "{{ .name }}" {
		  key                 			  = "{{ .name }}"
		  checksum_policy_type            = "{{ .checksum_policy_type }}"
		  snapshot_version_behavior       = "{{ .snapshot_version_behavior }}"
		  max_unique_snapshots            = {{ .max_unique_snapshots }}
		  handle_releases                 = {{ .handle_releases }}
		  handle_snapshots                = {{ .handle_snapshots }}
		  suppress_pom_consistency_checks = {{ .suppress_pom_consistency_checks }}
		}
	`

func TestAccLocalMavenRepository(t *testing.T) {

	_, fqrn, name := mkNames("maven-local", "artifactory_local_maven_repository")
	tempStruct := make(map[string]interface{})
	copyInterfaceMap(commonJavaParams, tempStruct)

	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = false

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, localJavaRepositoryBasic, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", tempStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", tempStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", tempStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", tempStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", tempStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", tempStruct["suppress_pom_consistency_checks"])),
				),
			},
		},
	})
}

func TestAccLocalGradleRepository(t *testing.T) {

	_, fqrn, name := mkNames("gradle-local", "artifactory_local_gradle_repository")
	tempStruct := make(map[string]interface{})
	copyInterfaceMap(commonJavaParams, tempStruct)

	tempStruct["name"] = name
	tempStruct["resource_name"] = strings.Split(fqrn, ".")[0]
	tempStruct["suppress_pom_consistency_checks"] = true

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: executeTemplate(fqrn, localJavaRepositoryBasic, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "checksum_policy_type", fmt.Sprintf("%s", tempStruct["checksum_policy_type"])),
					resource.TestCheckResourceAttr(fqrn, "snapshot_version_behavior", fmt.Sprintf("%s", tempStruct["snapshot_version_behavior"])),
					resource.TestCheckResourceAttr(fqrn, "max_unique_snapshots", fmt.Sprintf("%d", tempStruct["max_unique_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "handle_releases", fmt.Sprintf("%v", tempStruct["handle_releases"])),
					resource.TestCheckResourceAttr(fqrn, "handle_snapshots", fmt.Sprintf("%v", tempStruct["handle_snapshots"])),
					resource.TestCheckResourceAttr(fqrn, "suppress_pom_consistency_checks", fmt.Sprintf("%v", tempStruct["suppress_pom_consistency_checks"])),
				),
			},
		},
	})
}

func TestAccLocalGenericRepository(t *testing.T) {

	_, fqrn, name := mkNames("generic-local", "artifactory_local_generic_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := executeTemplate("TestAccLocalGenericRepository", `
		resource "artifactory_local_generic_repository" "{{ .name }}" {
		  key                 = "{{ .name }}"
		}
	`, params)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
		},
	})
}

func TestAccLocalNpmRepository(t *testing.T) {

	_, fqrn, name := mkNames("npm-local", "artifactory_local_npm_repository")
	params := map[string]interface{}{
		"name": name,
	}
	localRepositoryBasic := executeTemplate("TestAccLocalNpmRepository", `
		resource "artifactory_local_npm_repository" "{{ .name }}" {
		  key                 = "{{ .name }}"
		}
	`, params)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
		},
	})
}

func mkTestCase(repoType string, t *testing.T) (*testing.T, resource.TestCase) {
	name := fmt.Sprintf("terraform-local-%s-%d-full", repoType, rand.Int())
	resourceName := fmt.Sprintf("artifactory_local_%s_repository.%s", repoType, name)
	const localRepositoryConfigFull = `
		resource "artifactory_local_%s_repository" "%s" {
			key                             = "%s"
			description                     = "Test repo for %s"
			notes                           = "Test repo for %s"
		}
	`

	cfg := fmt.Sprintf(localRepositoryConfigFull, repoType, name, name, name, name)
	return t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(resourceName, testCheckRepo),
		Steps: []resource.TestStep{
			{
				Config: cfg,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "key", name),
					resource.TestCheckResourceAttr(resourceName, "package_type", repoType),
					resource.TestCheckResourceAttr(resourceName, "description", fmt.Sprintf("Test repo for %s", name)),
					resource.TestCheckResourceAttr(resourceName, "notes", fmt.Sprintf("Test repo for %s", name)),
				),
			},
		},
	}
}

func TestAccAllRepoTypesLocal(t *testing.T) {

	for _, repo := range repoTypesLikeGeneric {
		t.Run(fmt.Sprintf("TestLocal%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
			resource.Test(mkTestCase(repo, t))
		})
	}
}
