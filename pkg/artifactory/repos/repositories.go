package repos

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/retry"
	"github.com/jfrog/terraform-provider-artifactory/pkg/artifactory/util"
	"net/http"
)
const repositoriesEndpoint = "artifactory/api/repositories/"

func MkRepoCreate(unpack util.UnpackFunc, read schema.ReadContextFunc) schema.CreateContextFunc {

	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().AddRetryCondition(retry.OnMergeError).SetBody(repo).Put(repositoriesEndpoint + key)

		if err != nil {
			return diag.FromErr(err)
		}
		d.SetId(key)
		return read(ctx, d, m)
	}
}

func TestCheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	return CheckRepo(id, request.AddRetryCondition(retry.Never))
}
func RepoExists(d *schema.ResourceData, m interface{}) (bool, error) {
	_, err := CheckRepo(d.Id(), m.(*resty.Client).R().AddRetryCondition(retry.On400Error))
	return err == nil, err

}

func CheckRepo(id string, request *resty.Request) (*resty.Response, error) {
	// artifactory returns 400 instead of 404. but regardless, it's an error
	return request.Head(repositoriesEndpoint + id)
}
func MkRepoRead(pack util.PackFunc, construct util.Constructor) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo := construct()
		// repo must be a pointer
		resp, err := m.(*resty.Client).R().SetResult(repo).Get(repositoriesEndpoint + d.Id())

		if err != nil {
			if resp != nil && (resp.StatusCode() == http.StatusNotFound) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		return diag.FromErr(pack(repo, d))
	}
}

func MkRepoUpdate(unpack util.UnpackFunc, read schema.ReadContextFunc) schema.UpdateContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		repo, key, err := unpack(d)
		if err != nil {
			return diag.FromErr(err)
		}
		// repo must be a pointer
		_, err = m.(*resty.Client).R().AddRetryCondition(retry.OnMergeError).SetBody(repo).Post(repositoriesEndpoint + d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(key)
		return read(ctx, d, m)
	}
}

func DeleteRepo(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(*resty.Client).R().Delete(repositoriesEndpoint + d.Id())

	if err != nil && (resp != nil && resp.StatusCode() == http.StatusNotFound) {
		d.SetId("")
		return nil
	}
	return diag.FromErr(err)
}

func MkResourceSchema(skeema map[string]*schema.Schema, packer util.PackFunc, unpack util.UnpackFunc, constructor util.Constructor) *schema.Resource {
	var reader = MkRepoRead(packer, constructor)
	return &schema.Resource{
		CreateContext: MkRepoCreate(unpack, reader),
		ReadContext:   reader,
		UpdateContext: MkRepoUpdate(unpack, reader),
		DeleteContext: DeleteRepo,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: skeema,
	}
}








