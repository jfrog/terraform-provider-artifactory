package artifactory

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sort"
	"strings"
)

func resourceArtifactoryRemoteNpmRepository() *schema.Resource {

	npmRemoteSchema := mergeSchema(baseRemoteSchema, map[string]*schema.Schema{
		"list_remote_folder_items": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"mismatching_mime_types_override_list": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: commaSeperatedList,
			StateFunc: func(thing interface{}) string {
				fields := strings.Fields(thing.(string))
				sort.Strings(fields)
				return strings.Join(fields, ",")
			},
		},
	})
	type NpmRemoteRepository struct {
		RemoteRepositoryBaseParams
		ListRemoteFolderItems           bool   `json:"listRemoteFolderItems"`
		MismatchingMimeTypeOverrideList string `json:"mismatchingMimeTypesOverrideList"`
	}
	var unpack = func(s *schema.ResourceData) (interface{}, string, error) {
		d := &ResourceData{s}
		repo := NpmRemoteRepository{
			RemoteRepositoryBaseParams:      unpackBaseRemoteRepo(s, "npm"),
			ListRemoteFolderItems:           d.getBool("list_remote_folder_items", false),
			MismatchingMimeTypeOverrideList: d.getString("mismatching_mime_types_override_list", false),
		}
		return repo, repo.Id(), nil
	}

	return mkResourceSchema(npmRemoteSchema, inSchema(npmRemoteSchema), unpack, func() interface{} {
		return &NpmRemoteRepository{
			RemoteRepositoryBaseParams: RemoteRepositoryBaseParams{
				Rclass:              "remote",
				PackageType:         "npm",
				RemoteRepoLayoutRef: "npm-default",
				RepoLayoutRef:       "npm-default",
			},
		}
	})
}
