package replication

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
)

const (
	EndpointPath        = "artifactory/api/replications/"
	ReplicationEndpoint = "artifactory/api/replications/{repo_key}"
)

func resourceReplicationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	resp, err := m.(util.ProviderMetadata).Client.R().
		AddRetryCondition(client.RetryOnMergeError).
		Delete(EndpointPath + d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if resp.StatusCode() == http.StatusBadRequest || resp.StatusCode() == http.StatusNotFound {
		d.SetId("")
		return nil
	}

	if resp.IsError() {
		return diag.Errorf("%s", resp.String())
	}

	return nil
}

type repoConfiguration struct {
	Rclass string `json:"rclass"`
}

func verifyRepoRclass(repoKey, expectedRclass string, req *resty.Request) (bool, error) {
	var repoConfig repoConfiguration
	resp, err := req.
		SetResult(&repoConfig).
		Get("artifactory/api/repositories/" + repoKey)

	if err != nil {
		return false, fmt.Errorf("error getting repository configuration: %v", err)
	}

	if resp.IsError() {
		return false, fmt.Errorf("%s", resp.String())
	}

	return repoConfig.Rclass == expectedRclass, nil
}
