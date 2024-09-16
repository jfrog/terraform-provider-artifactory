package lifecycle_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/acctest"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

type artifactUploadResponse struct {
	Path      string                         `json:"path"`
	Checksums artifactUploadChecksumResponse `json:"checksums"`
}

type artifactUploadChecksumResponse struct {
	SHA256 string `json:"sha256"`
}

func uploadTestFile(t *testing.T, repoKey string) (string, string, error) {
	body, err := os.ReadFile("../../../../samples/multi1-3.7-20220310.233748-1.jar")
	if err != nil {
		return "", "", err
	}
	uri := fmt.Sprintf("/artifactory/%s/org/jfrog/test/multi1/3.7-SNAPSHOT/multi1-3.7-SNAPSHOT.jar", repoKey)

	var result artifactUploadResponse
	_, err = acctest.GetTestResty(t).R().
		SetHeader("Content-Type", "application/java-archive").
		SetBody(body).
		SetResult(&result).
		Put(uri)
	if err != nil {
		return "", "", err
	}

	return result.Path, result.Checksums.SHA256, nil
}

type Build struct {
	Version string        `json:"version"`
	Name    string        `json:"name"`
	Number  string        `json:"number"`
	Started string        `json:"started"`
	Modules []BuildModule `json:"modules"`
}

type BuildModule struct {
	ID        string                `json:"id"`
	Artifacts []BuildModuleArtifact `json:"artifacts"`
}

type BuildModuleArtifact struct {
	Type string `json:"type"`
	Name string `json:"name"`
	SHA1 string `json:"sha1"`
}

func uploadBuild(t *testing.T, name, number, projectKey string) error {
	build := Build{
		Version: "1.0.1",
		Name:    name,
		Number:  number,
		Started: time.Now().Format("2006-01-02T15:04:05.000Z0700"),
		Modules: []BuildModule{
			{
				ID: "org.jfrog.test:multi1:3.7-SNAPSHOT",
				Artifacts: []BuildModuleArtifact{
					{
						Type: "jar",
						Name: "multi1-3.7-SNAPSHOT.jar",
						SHA1: "f142780623aed30ba41d15a8db1ec24da8fd67e8",
					},
				},
			},
		},
	}

	restyClient := acctest.GetTestResty(t)

	req := restyClient.R()

	if projectKey != "" {
		req.SetQueryParam("project", projectKey)
	}

	res, err := req.
		SetBody(build).
		Put("artifactory/api/build")

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

func deleteBuild(t *testing.T, name, projectKey string) error {
	type Build struct {
		Name      string `json:"buildName"`
		BuildRepo string `json:"buildRepo"`
		DeleteAll bool   `json:"deleteAll"`
	}

	build := Build{
		Name:      name,
		DeleteAll: true,
	}

	restyClient := acctest.GetTestResty(t)

	req := restyClient.R()

	if projectKey != "" {
		build.BuildRepo = fmt.Sprintf("%s-build-info", projectKey)
	}

	res, err := req.
		SetBody(build).
		Post("artifactory/api/build/delete")

	if err != nil {
		return err
	}

	if res.IsError() {
		return fmt.Errorf("%s", res.String())
	}

	return nil
}

func TestAccReleaseBundleV2_full_aql(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}`

	testData := map[string]string{
		"name":         resourceName,
		"keypair_name": keyPairName,
		"repo_name":    repoName,
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
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source.aql", fmt.Sprintf("items.find({\"repo\": {\"$match\": \"%s\"}})", repoName)),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_artifacts(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	artifactPath, artifactChecksum, err := uploadTestFile(t, repoName)
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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "artifacts"

		source = {
			artifacts = [{
				path = "{{ .artifact_path }}"
				sha256 = "{{ .artifact_checksum }}"
			}]
		}
	}`

	testData := map[string]string{
		"name":              resourceName,
		"keypair_name":      keyPairName,
		"artifact_path":     fmt.Sprintf("%s%s", repoName, artifactPath),
		"artifact_checksum": artifactChecksum,
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
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source_type", "artifacts"),
					resource.TestCheckResourceAttr(fqrn, "source.artifacts.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "source.artifacts.0.path", testData["artifact_path"]),
					resource.TestCheckResourceAttr(fqrn, "source.artifacts.0.sha256", testData["artifact_checksum"]),
					resource.TestCheckResourceAttrSet(fqrn, "created"),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_builds(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

	repoName := fmt.Sprintf("test-repo-%d", testutil.RandomInt())
	acctest.CreateRepo(t, repoName, "local", "maven", true, true)

	_, _, err := uploadTestFile(t, repoName)
	if err != nil {
		t.Fatalf("failed to upload file: %s", err)
	}

	keyPairName := fmt.Sprintf("test-keypair-%d", testutil.RandomInt())
	buildName := fmt.Sprintf("test-build-%d", testutil.RandomInt())

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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "builds"

		source = {
			builds = [{
				name = "{{ .build_name }}"
				number = "{{ .build_number }}"
			}]
		}
	}`

	testData := map[string]string{
		"name":         resourceName,
		"keypair_name": keyPairName,
		"build_name":   buildName,
		"build_number": "1",
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full_builds", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			if err := uploadBuild(t, buildName, "1", ""); err != nil {
				t.Fatalf("failed to upload build: %s", err)
			}
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)

			if err := deleteBuild(t, buildName, ""); err != nil {
				return err
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source_type", "builds"),
					resource.TestCheckResourceAttr(fqrn, "source.builds.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "source.builds.0.name", testData["build_name"]),
					resource.TestCheckResourceAttr(fqrn, "source.builds.0.number", "1"),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_release_bundles(t *testing.T) {
	_, _, resourceName1 := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")
	_, fqrn, resourceName2 := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

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

	resource "artifactory_release_bundle_v2" "{{ .name1 }}" {
		name = "{{ .name1 }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}

	resource "artifactory_release_bundle_v2" "{{ .name2 }}" {
		name = "{{ .name2 }}"
		version = "2.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		skip_docker_manifest_resolution = true
		source_type = "release_bundles"

		source = {
			release_bundles = [{
				name = artifactory_release_bundle_v2.{{ .name1 }}.name
				version = artifactory_release_bundle_v2.{{ .name1 }}.version
			}]
		}
	}`

	testData := map[string]string{
		"name1":        resourceName1,
		"name2":        resourceName2,
		"keypair_name": keyPairName,
		"repo_name":    repoName,
	}

	config := util.ExecuteTemplate("TestAccReleaseBundleV2_full_builds", template, testData)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
		},
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			acctest.DeleteRepo(t, repoName)
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name2"]),
					resource.TestCheckResourceAttr(fqrn, "version", "2.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source_type", "release_bundles"),
					resource.TestCheckResourceAttr(fqrn, "source.release_bundles.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "source.release_bundles.0.name", testData["name1"]),
					resource.TestCheckResourceAttr(fqrn, "source.release_bundles.0.version", "1.0.0"),
				),
			},
		},
	})
}

func TestAccReleaseBundleV2_full_aql_with_project(t *testing.T) {
	_, fqrn, resourceName := testutil.MkNames("test-release-bundle-v2", "artifactory_release_bundle_v2")

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

	resource "artifactory_release_bundle_v2" "{{ .name }}" {
		name = "{{ .name }}"
		version = "1.0.0"
		keypair_name = artifactory_keypair.{{ .keypair_name }}.pair_name
		project_key = project.{{ .project_name }}.key
		skip_docker_manifest_resolution = true
		source_type = "aql"

		source = {
			aql = "items.find({\"repo\": {\"$match\": \"{{ .repo_name }}\"}})"
		}
	}`

	testData := map[string]string{
		"name":         resourceName,
		"keypair_name": keyPairName,
		"repo_name":    repoName,
		"project_name": projectName,
		"project_key":  projectKey,
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
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "version", "1.0.0"),
					resource.TestCheckResourceAttr(fqrn, "keypair_name", testData["keypair_name"]),
					resource.TestCheckResourceAttr(fqrn, "source.aql", fmt.Sprintf("items.find({\"repo\": {\"$match\": \"%s\"}})", repoName)),
				),
			},
		},
	})
}
