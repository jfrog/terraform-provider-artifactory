package artifactory

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVirtualRepository_basic(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryBasic = `
		resource "artifactory_virtual_repository" "%s" {
			key          = "%s"
			package_type = "maven"
			repositories = []
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryBasic, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
				),
			},
		},
	})
}

func TestAccVirtualRepository_reset_default_deployment_repo(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	localRepoName := fmt.Sprintf("%s-local", name)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryWithDefaultDeploymentRepo = `
		resource "artifactory_local_repository" "%[1]s" {
			key = "%[1]s"
			package_type = "maven"
		}

		resource "artifactory_virtual_repository" "%[2]s" {
			key          = "%[2]s"
			package_type = "maven"
			repositories = ["%[1]s"]
			default_deployment_repo = "%[1]s"
			depends_on = [artifactory_local_repository.%[1]s]
		}
	`
	const virtualRepositoryWithoutDefaultDeploymentRepo = `
		resource "artifactory_local_repository" "%[1]s" {
			key = "%[1]s"
			package_type = "maven"
		}

		resource "artifactory_virtual_repository" "%[2]s" {
			key          = "%[2]s"
			package_type = "maven"
			repositories = ["%[1]s"]
			depends_on = [artifactory_local_repository.%[1]s]
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryWithoutDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", ""),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryWithDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", localRepoName),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryWithoutDefaultDeploymentRepo, localRepoName, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "default_deployment_repo", ""),
				),
			},
		},
	})
}

func TestAccVirtualGoRepository_basic(t *testing.T) {
	_, fqrn, name := mkNames("foo", "artifactory_virtual_go_repository")
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_go_repository" "%s" {
		  key          = "%s"
		  repo_layout_ref = "go-default"
		  repositories = []
		  description = "A test virtual repo"
		  notes = "Internal description"
		  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
		  excludes_pattern = "com/google/**"
		  external_dependencies_enabled = true
		  external_dependencies_patterns = [
			"**/github.com/**",
			"**/go.googlesource.com/**"
		  ]
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "go"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_enabled", "true"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.0", "**/github.com/**"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.1", "**/go.googlesource.com/**"),
					resource.TestCheckResourceAttr(fqrn, "external_dependencies_patterns.#", "2"),
				),
			},
		},
	})
}

func TestAccVirtualConanRepository_basic(t *testing.T) {
	_, fqrn, name := mkNames("foo", "artifactory_virtual_conan_repository")
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_conan_repository" "%s" {
		  key          = "%s"
		  repo_layout_ref = "conan-default"
		  repositories = []
		  description = "A test virtual repo"
		  notes = "Internal description"
		  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
		  excludes_pattern = "com/google/**"
 		  retrieval_cache_period_seconds = 7100
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "conan"),
					resource.TestCheckResourceAttr(fqrn, "retrieval_cache_period_seconds", "7100"),
				),
			},
		},
	})
}

func TestAccVirtualGenericRepository_basic(t *testing.T) {
	_, fqrn, name := mkNames("foo", "artifactory_virtual_generic_repository")
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_generic_repository" "%s" {
		  key          = "%s"
		  repo_layout_ref = "simple-default"
		  repositories = []
		  description = "A test virtual repo"
		  notes = "Internal description"
		  includes_pattern = "com/jfrog/**,cloud/jfrog/**"
		  excludes_pattern = "com/google/**"
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "generic"),
				),
			},
		},
	})
}

func TestAccVirtualMavenRepository_basic(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_maven_repository.%s", name)
	var virtualRepositoryBasic = fmt.Sprintf(`
		resource "artifactory_virtual_maven_repository" "%s" {
			key          = "%s"
			repo_layout_ref = "maven-2-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			force_maven_authentication = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`, name, name)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				// we check to make sure some of the base params are picked up
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "force_maven_authentication", "true"),
					// to test key pair, we'd have to be able to create them on the fly and we currently can't.
					resource.TestCheckResourceAttr(fqrn, "key_pair", ""),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
				),
			},
		},
	})
}

func TestAccVirtualRpmRepository(t *testing.T) {
	_, fqrn, name := mkNames("virtual-rpm-repo", "artifactory_virtual_rpm_repository")
	kpId, kpFqrn, kpName := mkNames("some-keypair1-", "artifactory_keypair")
	kpId2, kpFqrn2, kpName2 := mkNames("some-keypair2-", "artifactory_keypair")
	virtualRepositoryBasic := executeTemplate("keypair", `
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
		resource "artifactory_virtual_rpm_repository" "{{ .repo_name }}" {
			key 	              = "{{ .repo_name }}"
			primary_keypair_ref   = artifactory_keypair.{{ .kp_name }}.pair_name
			secondary_keypair_ref = artifactory_keypair.{{ .kp_name2 }}.pair_name

			depends_on = [
				artifactory_keypair.{{ .kp_name }},
				artifactory_keypair.{{ .kp_name2 }},
			]
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
				Config: virtualRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "rpm"),
					resource.TestCheckResourceAttr(fqrn, "primary_keypair_ref", kpName),
					resource.TestCheckResourceAttr(fqrn, "secondary_keypair_ref", kpName2),
				),
			},
		},
	})
}

func TestAccVirtualRepository_update(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryUpdateBefore = `
		resource "artifactory_virtual_repository" "%s" {
			key          = "%s"
			description  = "Before"
			package_type = "maven"
			repositories = []
		}
	`
	const virtualRepositoryUpdateAfter = `
		resource "artifactory_virtual_repository" "%s" {
			key          = "%s"
			description  = "After"
			package_type = "maven"
			repositories = []
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryUpdateBefore, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "Before"),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(virtualRepositoryUpdateAfter, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "description", "After"),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
				),
			},
		},
	})
}
func TestAllPackageTypes(t *testing.T) {
	for _, repo := range repoTypesSupported {
		if repo != "nuget" { // this requires special testing
			t.Run(fmt.Sprintf("TestVirtual%sRepo", strings.Title(strings.ToLower(repo))), func(t *testing.T) {
				// NuGet Repository configuration is missing mandatory field downloadContextPath
				resource.Test(mkVirtualTestCase(repo, t))
			})
		}
	}
}

func mkVirtualTestCase(repo string, t *testing.T) (*testing.T, resource.TestCase) {
	id := randomInt()
	name := fmt.Sprintf("%s%d", repo, id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_repository" "%s" {
			key = "%s"
			package_type = "%s"
			repo_layout_ref = "maven-1-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`
	return t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name, repo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", repo),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "maven-1-default"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "description", "A test virtual repo"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal description"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "com/jfrog/**,cloud/jfrog/**"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "com/google/**"),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
				),
			},
		},
	}
}

func TestNugetPackageCreationFull(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_repository" "%s" {
			key = "%s"
			package_type = "nuget"
			repo_layout_ref = "nuget-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
			force_nuget_authentication	= true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "nuget"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "nuget-default"),
					resource.TestCheckResourceAttr(fqrn, "force_nuget_authentication", "true"),
				),
			},
		},
	})

}
func TestAccVirtualRepository_full(t *testing.T) {
	id := randomInt()
	name := fmt.Sprintf("foo%d", id)
	fqrn := fmt.Sprintf("artifactory_virtual_repository.%s", name)
	const virtualRepositoryFull = `
		resource "artifactory_virtual_repository" "%s" {
			key = "%s"
			package_type = "maven"
			repo_layout_ref = "maven-1-default"
			repositories = []
			description = "A test virtual repo"
			notes = "Internal description"
			includes_pattern = "com/jfrog/**,cloud/jfrog/**"
			excludes_pattern = "com/google/**"
			artifactory_requests_can_retrieve_remote_artifacts = true
			pom_repository_references_cleanup_policy = "discard_active_reference"
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      verifyDeleted(fqrn, testCheckRepo),
		ProviderFactories: testAccProviders,

		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(virtualRepositoryFull, name, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "package_type", "maven"),
					resource.TestCheckResourceAttr(fqrn, "repo_layout_ref", "maven-1-default"),
					resource.TestCheckResourceAttr(fqrn, "repositories.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "description", "A test virtual repo"),
					resource.TestCheckResourceAttr(fqrn, "notes", "Internal description"),
					resource.TestCheckResourceAttr(fqrn, "includes_pattern", "com/jfrog/**,cloud/jfrog/**"),
					resource.TestCheckResourceAttr(fqrn, "excludes_pattern", "com/google/**"),
					resource.TestCheckResourceAttr(fqrn, "pom_repository_references_cleanup_policy", "discard_active_reference"),
				),
			},
		},
	})
}

func TestAccVirtualGenericRepositoryWithProjectAttributesGH318(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", rand.Intn(100000000))
	projectEnv := randSelect("DEV", "PROD").(string)
	repoName := fmt.Sprintf("%s-generic-virtual", projectKey)

	_, fqrn, name := mkNames(repoName, "artifactory_virtual_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
		"projectEnv": projectEnv,
	}
	virtualRepositoryBasic := executeTemplate("TestAccVirtualGenericRepository", `
		resource "artifactory_virtual_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "{{ .projectKey }}"
	 	  project_environments = ["{{ .projectEnv }}"]
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			createProject(t, projectKey)
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			deleteProject(t, projectKey)
			return testCheckRepo(id, request)
		}),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: virtualRepositoryBasic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "key", name),
					resource.TestCheckResourceAttr(fqrn, "project_key", projectKey),
					resource.TestCheckResourceAttr(fqrn, "project_environments.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "project_environments.0", projectEnv),
				),
			},
		},
	})
}

func TestAccVirtualRepositoryWithInvalidProjectKeyGH318(t *testing.T) {

	rand.Seed(time.Now().UnixNano())
	projectKey := fmt.Sprintf("t%d", rand.Intn(100000000))
	repoName := fmt.Sprintf("%s-generic-virtual", projectKey)

	_, fqrn, name := mkNames(repoName, "artifactory_virtual_generic_repository")

	params := map[string]interface{}{
		"name":       name,
		"projectKey": projectKey,
	}
	virualRepositoryBasic := executeTemplate("TestAccVirtualGenericRepository", `
		resource "artifactory_virtual_generic_repository" "{{ .name }}" {
		  key                  = "{{ .name }}"
	 	  project_key          = "invalid-project-key"
		}
	`, params)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			createProject(t, projectKey)
		},
		CheckDestroy: verifyDeleted(fqrn, func(id string, request *resty.Request) (*resty.Response, error) {
			deleteProject(t, projectKey)
			return testCheckRepo(id, request)
		}),
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: virualRepositoryBasic,
				ExpectError: regexp.MustCompile(".*project_key must be 3 - 10 lowercase alphanumeric characters"),
			},
		},
	})
}
