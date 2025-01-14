package local_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func TestAccLocalDebianRepository(t *testing.T) {
	_, fqrn, name := testutil.MkNames("local-debian-repo", "artifactory_local_debian_repository")
	kpId, kpFqrn, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")
	localRepositoryBasic := util.ExecuteTemplate("keypair", `
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
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
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
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_local_debian_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = ["bz2","lzma","xz"]
			trivial_layout = true
			ddeb_supported = true
		}
	`, map[string]interface{}{
		"kp_id":     kpId,
		"kp_name":   kpName,
		"kp_id2":    kpId2,
		"kp_name2":  kpName2,
		"repo_name": name,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: acctest.CompositeCheckDestroy(
			acctest.VerifyDeleted(t, fqrn, "", acctest.CheckRepo),
			acctest.VerifyDeleted(t, kpFqrn, "pair_name", security.VerifyKeyPair),
			acctest.VerifyDeleted(t, kpFqrn2, "pair_name", security.VerifyKeyPair),
		),
		Steps: []resource.TestStep{
			{
				Config: localRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.0", "bz2"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.1", "lzma"),
					resource.TestCheckResourceAttr(fqrn, "index_compression_formats.2", "xz"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", func() string {
						r, _ := repository.GetDefaultRepoLayoutRef("local", repository.DebianPackageType)
						return r
					}()), //Check to ensure repository layout is set as per default even when it is not passed.
					resource.TestCheckResourceAttr(fqrn, "trivial_layout", "true"),
					resource.TestCheckResourceAttr(fqrn, "ddeb_supported", "true"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateCheck:  validator.CheckImportState(name, "key"),
			},
		},
	})
}

func TestAccLocalDebianRepository_UpgradeFromSDKv2(t *testing.T) {
	_, fqrn, name := testutil.MkNames("local-debian-repo", "artifactory_local_debian_repository")
	kpId, _, kpName := testutil.MkNames("some-keypair1", "artifactory_keypair")
	kpId2, _, kpName2 := testutil.MkNames("some-keypair2", "artifactory_keypair")

	config := util.ExecuteTemplate("keypair", `
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
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
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
			lifecycle {
				ignore_changes = [
					private_key,
					passphrase,
				]
			}
		}

		resource "artifactory_local_debian_repository" "{{ .repo_name }}" {
			key = "{{ .repo_name }}"
			primary_keypair_ref = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name
			index_compression_formats = ["bz2","lzma","xz"]
			trivial_layout = true
		}
	`, map[string]interface{}{
		"kp_id":     kpId,
		"kp_name":   kpName,
		"kp_id2":    kpId2,
		"kp_name2":  kpName2,
		"repo_name": name,
	}) // we use randomness so that, in the case of failure and dangle, the next test can run without collision

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"artifactory": {
						VersionConstraint: "12.8.0",
						Source:            "jfrog/artifactory",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", name),
					resource.TestCheckResourceAttr(fqrn, "key", name),
				),
			},
			{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
