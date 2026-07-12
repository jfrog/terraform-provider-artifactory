// Copyright (c) JFrog Ltd. (2025)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	EndpointPath             = "artifactory/api/replications/"
	ReplicationEndpoint      = "artifactory/api/replications/{repo_key}"
	MultiReplicationEndpoint = "artifactory/api/replications/multiple/{repo_key}"
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
