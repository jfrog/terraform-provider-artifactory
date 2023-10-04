package security_test

import (
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/acctest"
	datasource "github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/datasource/security"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v9/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-shared/testutil"
	utilsdk "github.com/jfrog/terraform-provider-shared/util/sdk"
)

func deletePermissionTarget(t *testing.T, name string) error {
	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().Delete(security.PermissionsEndPoint + name)

	return err
}

func createPermissionTarget(targetName string, userName string, t *testing.T) {
	acctest.CreateUserUpdatable(t, userName, "terraform@email.com")

	actions := datasource.Actions{
		Users:  map[string][]string{userName: {"read", "write"}},
		Groups: map[string][]string{"readers": {"read"}},
	}
	repoTarget := datasource.PermissionTargetSection{
		IncludePatterns: []string{"foo/**"},
		ExcludePatterns: []string{"bar/**"},
		Repositories:    []string{"example-repo-local"},
		Actions:         &actions,
	}
	buildTarget := datasource.PermissionTargetSection{
		IncludePatterns: []string{"foo/**"},
		ExcludePatterns: []string{"bar/**"},
		Repositories:    []string{"artifactory-build-info"},
		Actions:         nil,
	}
	releaseBundleTarget := datasource.PermissionTargetSection{
		IncludePatterns: []string{"foo/**"},
		ExcludePatterns: []string{"bar/**"},
		Repositories:    []string{"release-bundles"},
		Actions:         nil,
	}

	permissionTarget := datasource.PermissionTargetParams{
		Name:          targetName,
		Repo:          &repoTarget,
		Build:         &buildTarget,
		ReleaseBundle: &releaseBundleTarget,
	}

	restyClient := acctest.GetTestResty(t)
	postPermissionTarget := security.PermissionsEndPoint + permissionTarget.Name
	if _, err := restyClient.R().AddRetryCondition(repository.Retry400).SetBody(permissionTarget).Post(postPermissionTarget); err != nil {
		t.Fatal(err)
	}

	log.Printf("Create permission target #{permissionTarget.Name} done.")
}

func TestAccDataSourcePermissionTarget_full(t *testing.T) {
	_, fqrn, name := testutil.MkNames("test-perm", "data.artifactory_permission_target")

	temp := `
  data "artifactory_permission_target" "{{ .permission_name }}" {
    name = "{{ .permission_name }}"
  }`

	tempStruct := map[string]string{
		"permission_name": name,
	}

	_, _, userName := testutil.MkNames("test-user", "artifactory_managed_user")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createPermissionTarget(name, userName, t)
		},
		ProtoV5ProviderFactories: acctest.ProtoV5MuxProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			err := deletePermissionTarget(t, name)
			_ = acctest.DeleteUser(t, userName)
			return err
		},
		Steps: []resource.TestStep{
			{
				Config: utilsdk.ExecuteTemplate(fqrn, temp, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.actions.0.users.0.%", "2"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.repositories.0", "example-repo-local"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.includes_pattern.0", "foo/**"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.excludes_pattern.0", "bar/**"),
					resource.TestCheckResourceAttr(fqrn, "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "build.0.repositories.0", "artifactory-build-info"),
					resource.TestCheckResourceAttr(fqrn, "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "build.0.includes_pattern.0", "foo/**"),
					resource.TestCheckResourceAttr(fqrn, "build.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "build.0.excludes_pattern.0", "bar/**"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.repositories.0", "release-bundles"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.includes_pattern.0", "foo/**"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.excludes_pattern.0", "bar/**")),
			},
		},
	})
}
