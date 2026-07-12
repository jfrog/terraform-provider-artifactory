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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-shared/validator"
)

func resourcePushReplicationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_push_replication deprecated. Use artifactory_local_repository_multi_replication instead")
}

func resourcePushReplicationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_push_replication deprecated. Use artifactory_local_repository_multi_replication instead")
}

func resourcePushReplicationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.Errorf("artifactory_push_replication deprecated. Use artifactory_local_repository_multi_replication instead")
}

var pushReplicationSchema = map[string]*schema.Schema{
	"url": {
		Type:             schema.TypeString,
		Required:         true,
		ForceNew:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IsURLWithHTTPorHTTPS),
	},
	"socket_timeout_millis": {
		Type:             schema.TypeInt,
		Optional:         true,
		Computed:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.IntAtLeast(0)),
	},
	"username": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Username for push replication",
	},
	"password": {
		Type:             schema.TypeString,
		Required:         true,
		Sensitive:        true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      "Password for push replication",
	},
	"enabled": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"sync_deletes": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"sync_properties": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"sync_statistics": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
	"path_prefix": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"proxy": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Proxy key from Artifactory Proxies setting",
	},
	"check_binary_existence_in_filestore": {
		Type:     schema.TypeBool,
		Optional: true,
		Description: "When true, enables distributed checksum storage. For more information, see " +
			"[Optimizing Repository Replication with Checksum-Based Storage](https://www.jfrog.com/confluence/display/JFROG/Repository+Replication#RepositoryReplication-OptimizingRepositoryReplicationUsingStorageLevelSynchronizationOptions).",
	},
}

var pushRepMultipleSchema = map[string]*schema.Schema{
	"cron_exp": {
		Type:             schema.TypeString,
		Required:         true,
		ValidateDiagFunc: validator.CronLength,
		Description:      "Cron expression to control the operation frequency.",
	},
	"repo_key": {
		Type:     schema.TypeString,
		Required: true,
	},
	"enable_event_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: "When set, each event will trigger replication of the artifacts changed in this event. This can be any type of event on artifact, e.g. add, deleted or property change. Default value is `false`.",
	},
	"replications": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: pushReplicationSchema,
		},
	},
}

func ResourceArtifactoryPushReplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePushReplicationCreate,
		ReadContext:   resourcePushReplicationRead,
		UpdateContext: resourcePushReplicationUpdate,
		DeleteContext: resourceReplicationDelete,

		Schema:             pushRepMultipleSchema,
		Description:        "Add or replace multiple replication configurations for given repository key. Supported by local repositories. Artifactory Enterprise license is required.",
		DeprecationMessage: "This resource is replaced by `artifactory_local_repository_multi_replication` for clarity. All the attributes are identical, please consider the migration.",
	}
}
