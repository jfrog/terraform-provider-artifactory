package security_test

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/repository"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/security"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory/resource/user"
	"github.com/jfrog/terraform-provider-shared/test"
	"github.com/jfrog/terraform-provider-shared/util"
	"log"
	"testing"
)

func deletePermissionTarget(t *testing.T, name string) error {
	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().Delete(security.PermissionsEndPoint + name)

	return err
}

// TODO: Where can we put this user code
func deleteUser(t *testing.T, name string) error {
	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().Delete(user.UsersEndpointPath + name)

	return err
}

func createUserUpdatable(t *testing.T, name string, email string) {
	userObj := user.User{
		Name:                     name,
		Email:                    email,
		Password:                 "Lizard123!",
		Admin:                    false,
		ProfileUpdatable:         true,
		DisableUIAccess:          false,
		InternalPasswordDisabled: false,
		Groups:                   []string{"readers"},
	}

	restyClient := acctest.GetTestResty(t)
	_, err := restyClient.R().SetBody(userObj).Put(user.UsersEndpointPath + name)

	if err != nil {
		t.Fatal(err)
	}
}

func createPermissionTarget(targetName string, userName string, t *testing.T) {
	createUserUpdatable(t, userName, "terraform@email.com")

	actions := security.Actions{
		Users:  map[string][]string{"terraform": {"read", "write"}},
		Groups: map[string][]string{"readers": {"read"}},
	}
	repoTarget := security.PermissionTargetSection{
		IncludePatterns: []string{"foo/**"},
		ExcludePatterns: []string{"bar/**"},
		Repositories:    []string{"example-repo-local"},
		Actions:         &actions,
	}
	buildTarget := security.PermissionTargetSection{
		IncludePatterns: []string{"foo/**"},
		ExcludePatterns: []string{"bar/**"},
		Repositories:    []string{"artifactory-build-info"},
		Actions:         nil,
	}
	releaseBundleTarget := security.PermissionTargetSection{
		IncludePatterns: []string{"foo/**"},
		ExcludePatterns: []string{"bar/**"},
		Repositories:    []string{"release-bundles"},
		Actions:         nil,
	}

	permissionTarget := security.PermissionTargetParams{
		Name:          targetName,
		Repo:          &repoTarget,
		Build:         &buildTarget,
		ReleaseBundle: &releaseBundleTarget,
	}

	restyClient := acctest.GetTestResty(t)
	if _, err := restyClient.R().AddRetryCondition(repository.Retry400).SetBody(permissionTarget).Post(security.PermissionsEndPoint + permissionTarget.Name); err != nil {
		t.Fatal(err)
	}

	log.Printf("Create permission target #{permissionTarget.Name} done.")
}

func TestAccDataSourcePermissionTarget_full(t *testing.T) {
	_, tempFqrn, name := test.MkNames("test-perm", "artifactory_permission_target")

	temp := `
  data "artifactory_permission_target" "{{ .permission_name }}" {
    name = "{{ .permission_name }}"
  }`

	tempStruct := map[string]string{
		"permission_name": name,
	}

	fqrn := "data." + tempFqrn
	userName := "terraform"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			createPermissionTarget(name, userName, t)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy: func(state *terraform.State) error {
			err := deletePermissionTarget(t, name)
			err2 := deleteUser(t, userName)
			// TODO: Figure out wtf this is. How do I turn 2 errors into 1.
			if err != nil {
				return err
			} else {
				return err2
			}
		},
		Steps: []resource.TestStep{
			{
				Config: util.ExecuteTemplate(fqrn, temp, tempStruct),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", name),
					resource.TestCheckResourceAttr(fqrn, "repo.0.actions.0.users.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.actions.0.groups.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "repo.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "build.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "build.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "build.0.excludes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.repositories.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.includes_pattern.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "release_bundle.0.excludes_pattern.#", "1"),
				),
			},
		},
	})
}
