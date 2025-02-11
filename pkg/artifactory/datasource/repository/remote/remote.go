package remote

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jfrog/terraform-provider-artifactory/v12/pkg/artifactory/resource/repository/remote"
)

var getSchema = func(schemas map[int16]map[string]*schema.Schema) map[string]*schema.Schema {
	s := schemas[remote.CurrentSchemaVersion]

	s["url"].Required = false
	s["url"].Optional = true

	return s
}

var VcsRemoteRepoSchemaSDKv2 = map[string]*schema.Schema{
	"vcs_git_provider": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          "GITHUB",
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"GITHUB", "BITBUCKET", "OLDSTASH", "STASH", "ARTIFACTORY", "CUSTOM"}, false)),
		Description:      `Artifactory supports proxying the following Git providers out-of-the-box: GitHub or a remote Artifactory instance. Default value is "GITHUB".`,
	},
	"vcs_git_download_url": {
		Type:             schema.TypeString,
		Optional:         true,
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
		Description:      `This attribute is used when vcs_git_provider is set to 'CUSTOM'. Provided URL will be used as proxy.`,
	},
}
