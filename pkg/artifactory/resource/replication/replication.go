package replication

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/jfrog/terraform-provider-shared/client"
	"github.com/jfrog/terraform-provider-shared/util"
)

const EndpointPath = "artifactory/api/replications/"

var replicationSchemaEnableEventReplication = map[string]*schema.Schema{
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
	},
}

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

func getRepositoryRclass(repoKey string, m interface{}) (string, error) {
	repoConfig := repoConfiguration{}
	resp, err := m.(util.ProviderMetadata).Client.R().
		SetResult(&repoConfig).
		Get("artifactory/api/repositories/" + repoKey)
	if err != nil {
		return "", err
	}
	if resp.IsError() {
		return "", fmt.Errorf("%s", resp.String())
	}

	return repoConfig.Rclass, err
}

func verifyRepoRclass(repoKey string, expectedRclass string, m interface{}) (bool, error) {
	rclass, err := getRepositoryRclass(repoKey, m)
	if err != nil {
		return false, fmt.Errorf("error getting repository configuration: %v", err)
	}
	if rclass == expectedRclass {
		return true, nil
	}
	return false, nil
}
