package artifactory

import (
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"net/http"
	"regexp"
)

const repositoriesEndpoint = "artifactory/api/repositories/"

type ReadFunc func(d *schema.ResourceData, m interface{}) error

// Constructor Must return a pointer to a struct. When just returning a struct, it somehow converts to a map
type Constructor func() interface{}

// UnpackFunc must return a pointer to a struct and the resource id
type UnpackFunc func(s *schema.ResourceData) (interface{}, string)

type PackFunc func(repo interface{}, d *schema.ResourceData) error


func mkRepoCreate(unpack UnpackFunc, read ReadFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		repo, key := unpack(d)
		// repo must be a pointer
		_, err := m.(*resty.Client).R().AddRetryCondition(func(response *resty.Response, _r error) bool {
			return regexp.MustCompile(".*Could not merge and save new descriptor.*").MatchString(string(response.Body()[:]))
		}).SetBody(repo).Put(repositoriesEndpoint + key)

		if err != nil {
			return err
		}
		d.SetId(key)
		return read(d, m)
	}
}

func mkRepoRead(pack PackFunc, construct Constructor) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		repo := construct()
		// repo must be a pointer
		resp, err := m.(*resty.Client).R().SetResult(repo).Get(repositoriesEndpoint + d.Id())

		if err != nil {
			if resp != nil && (resp.StatusCode() == http.StatusNotFound) {
				d.SetId("")
				return nil
			}
			return err
		}
		return pack(repo, d)
	}
}

func mkRepoUpdate(unpack UnpackFunc, read ReadFunc) func(d *schema.ResourceData, m interface{}) error {
	return func(d *schema.ResourceData, m interface{}) error {
		repo, key := unpack(d)
		// repo must be a pointer
		_, err := m.(*resty.Client).R().SetBody(repo).Post(repositoriesEndpoint + d.Id())
		if err != nil {
			return err
		}

		d.SetId(key)
		return read(d, m)
	}
}

func deleteRepo(d *schema.ResourceData, m interface{}) error {
	resp, err := m.(*resty.Client).R().Delete(repositoriesEndpoint + d.Id())

	if err != nil && (resp != nil && resp.StatusCode() == http.StatusNotFound) {
		d.SetId("")
		return nil
	}
	return err
}

func checkRepo(id string, m interface{}) (bool, error) {
	_, err := m.(*resty.Client).R().AddRetryCondition(func(response *resty.Response, err error) bool {
		return response.StatusCode() == 400
	}).Head(repositoriesEndpoint + id)
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return err == nil, err
}

func repoExists(d *schema.ResourceData, m interface{}) (bool, error) {
	return checkRepo(d.Id(), m)
}

var repoTypeValidator = validation.StringInSlice(repoTypesSupported, false)

var repoKeyValidator = validation.All(
	validation.StringDoesNotMatch(regexp.MustCompile("^[0-9].*"), "repo key cannot start with a number"),
	validation.StringDoesNotContainAny(" !@#$%^&*()_+={}[]:;<>,/?~`|\\"),
)

var repoTypesSupported = []string{
	"alpine",
	"bower",
	//"cargo", // not supported
	"chef",
	"cocoapods",
	"composer",
	"conan",
	"conda",
	"cran",
	"debian",
	"docker",
	"gems",
	"generic",
	"gitlfs",
	"go",
	"gradle",
	"helm",
	"ivy",
	"maven",
	"npm",
	"nuget",
	"opkg",
	"p2",
	"puppet",
	"pypi",
	"rpm",
	"sbt",
	"vagrant",
	"vcs",
}