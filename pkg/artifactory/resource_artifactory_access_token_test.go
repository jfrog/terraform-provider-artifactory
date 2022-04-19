package artifactory_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/acctest"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/artifactory"
	"github.com/jfrog/terraform-provider-artifactory/v6/pkg/utils"
)

func TestAccAccessTokenAudienceBad(t *testing.T) {
	const audienceBad = `
		resource "artifactory_user" "existinguser" {
			name  = "existinguser"
			email = "existinguser@a.com"
			admin = false
			groups = ["readers"]
			password = "Passw0rd!"
		}

		resource "artifactory_access_token" "foobar" {
			end_date_relative = "1s"
			username = artifactory_user.existinguser.name
			audience = "bad"
			refreshable = true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenNotCreated("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config:      audienceBad,
				ExpectError: regexp.MustCompile("audience can contain only service IDs of Artifactory servers"),
			},
		},
	})
}

func TestAccAccessTokenAudienceGood(t *testing.T) {
	const audienceGood = `
		resource "artifactory_user" "existinguser" {
			name  = "existinguser"
			email = "existinguser@a.com"
			admin = false
			groups = ["readers"]
			password = "Passw0rd!"
		}

		resource "artifactory_access_token" "foobar" {
			end_date_relative = "1s"
			username = artifactory_user.existinguser.name
			audience = "jfrt@*"
			refreshable = true
		}
	`
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenDestroy("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config: audienceGood,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "access_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "username", "existinguser"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "end_date_relative", "1s"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "end_date"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refreshable", "true"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "refresh_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.#", "0"),
				),
			},
		},
	})
}

const existingUser = `
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
    email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date_relative = "1s"
	username = artifactory_user.existinguser.name
}
`

func TestAccAccessTokenExistingUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenDestroy("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config: existingUser,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "access_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "username", "existinguser"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "end_date_relative", "1s"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "end_date"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refreshable", "false"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refresh_token", ""),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.#", "0"),
				),
			},
		},
	})
}

func fixedDateGood() string {
	// Create a "fixed date" in the future

	date := time.Now().Add(time.Second * time.Duration(10)).Format(time.RFC3339)
	return fmt.Sprintf(`
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
    email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date = "%s"
	username = artifactory_user.existinguser.name
}
`, date)
}

func TestAccAccessTokenFixedDateGood(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenDestroy("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config: fixedDateGood(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "access_token"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "end_date"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "username", "existinguser"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refreshable", "false"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refresh_token", ""),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.#", "0"),
				),
			},
		},
	})
}

var fixedDateBad = `
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
    email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date = "2020-01-01T00:00:00+11:00"
	username = artifactory_user.existinguser.name
}
`

func TestAccAccessTokenFixedDateBad(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenNotCreated("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config:      fixedDateBad,
				ExpectError: regexp.MustCompile("end date must be in the future, but is"),
			},
		},
	})
}

// I couldn't find a way to retrieve the instance_id for artifactory via the go library.
// So, we expect this to fail
const adminToken = `
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
    email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date_relative = "1s"
	username = artifactory_user.existinguser.name

	admin_token {
		instance_id = "blah"
	}
}
`

func TestAccAccessTokenAdminToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenNotCreated("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config:      adminToken,
				ExpectError: regexp.MustCompile("Admin can create token with admin privileges only on this Artifactory instance"),
			},
		},
	})
}

const refreshableToken = `
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
    email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date_relative = "1s"
	refreshable = true
	username = artifactory_user.existinguser.name
}
`

func TestAccAccessTokenRefreshableToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenDestroy("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config: refreshableToken,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "access_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "username", "existinguser"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "end_date_relative", "1s"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "end_date"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refreshable", "true"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "refresh_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.#", "0"),
				),
			},
		},
	})
}

const missingUserBad = `
resource "artifactory_access_token" "foobar" {
	end_date_relative = "1s"
	username = "missing-user"
	groups = [
	]
}
`

func TestAccAccessTokenMissingUserBad(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenNotCreated("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config:      missingUserBad,
				ExpectError: regexp.MustCompile("you must specify at least 1 group when creating a token for a non-existant user"),
			},
		},
	})
}

const missingUserGood = `
resource "artifactory_access_token" "foobar" {
	end_date_relative = "1s"
	username = "missing-user"
	groups = [
		"readers",
	]
}
`

func TestAccAccessTokenMissingUserGood(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenDestroy("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config: missingUserGood,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "access_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "username", "missing-user"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "end_date_relative", "1s"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "end_date"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refreshable", "false"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refresh_token", ""),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.#", "1"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.0", "readers"),
				),
			},
		},
	})
}

const missingGroup = `
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
    email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date_relative = "1s"
	username = artifactory_user.existinguser.name
	groups = [
		"missing-group",
	]
}
`

func TestAccAccessTokenMissingGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenNotCreated("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config:      missingGroup,
				ExpectError: regexp.MustCompile("group must exist in artifactory"),
			},
		},
	})
}

const wildcardGroupGood = `
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
  email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date_relative = "1s"
	username = artifactory_user.existinguser.name
	groups = [
		"*",
	]
}
`

func TestAccAccessTokenWildcardGroupGood(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenDestroy("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config: wildcardGroupGood,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "access_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "username", "existinguser"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "end_date_relative", "1s"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "end_date"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refreshable", "false"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refresh_token", ""),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.#", "1"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.0", "*"),
				),
			},
		},
	})
}

const nonExpiringToken = `
resource "artifactory_user" "existinguser" {
	name  = "existinguser"
    email = "existinguser@a.com"
	admin = false
	groups = ["readers"]
	password = "Passw0rd!"
}

resource "artifactory_access_token" "foobar" {
	end_date_relative = "0s"
	username = artifactory_user.existinguser.name
}
`

func TestAccAccessTokenNonExpiringToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAccessTokenDestroy("artifactory_access_token.foobar"),
		Steps: []resource.TestStep{
			{
				Config: nonExpiringToken,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "access_token"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "username", "existinguser"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "end_date_relative", "0s"),
					resource.TestCheckResourceAttrSet("artifactory_access_token.foobar", "end_date"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refreshable", "false"),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "refresh_token", ""),
					resource.TestCheckResourceAttr("artifactory_access_token.foobar", "groups.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAccessTokenDestroy(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]

		if !ok {
			return fmt.Errorf("err: Resource id[%s] not found", id)
		}

		// We need to try to auth with the token and check that it errors
		// Thus, token has really been "destroyed"

		// This is more complicated when token has TTL, as Artifactory **does not** allow you to revoke a token that has a TTL.
		// https://www.jfrog.com/confluence/display/JFROG/Access+Tokens#AccessTokens-ViewingandRevokingTokens
		// https://www.jfrog.com/jira/browse/RTFACT-15293

		// Create a new client to auth to Artifactory
		// We want to check that the token cannot authenticate
		url := os.Getenv("ARTIFACTORY_URL")

		resty, err := utils.BuildResty(url, "")
		if err != nil {
			return err
		}
		accessToken := rs.Primary.Attributes["access_token"]
		resty, err = utils.AddAuthToResty(resty, "", accessToken)
		if err != nil {
			return err
		}
		if resp, err := resty.R().Get("artifactory/api/system/ping"); err != nil {
			if resp == nil {
				return fmt.Errorf("no response returned for testAccCheckAccessTokenDestroy")
			}
			if resp.StatusCode() == http.StatusUnauthorized {
				return nil
			}
			return fmt.Errorf("failed to ping server. Got %s", err)
		}
		return nil
	}
}

func testAccCheckAccessTokenNotCreated(id string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[id]

		if ok {
			return fmt.Errorf("err: Resource id[%s] found, but should not exist", id)
		}

		return nil
	}
}

func TestTokenOptsToValuesNonExpiring(t *testing.T) {
	tOpts := artifactory.AccessTokenOptions{Username: "existinguser", ExpiresIn: 0}
	values, _ := artifactory.TokenOptsToValues(tOpts)
	if values.Get("expires_in") != "0" {
		t.Error("`expires_in` not correctly set when creating non-expiring tokens")
	}
}
