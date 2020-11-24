package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	artifactoryold "github.com/atlassian/go-artifactory/v2/artifactory"
	"github.com/hashicorp/terraform/helper/schema"
	xrayutils "github.com/jfrog/jfrog-client-go/xray/services/utils"
)

func resourceArtifactoryWatch() *schema.Resource {
	repoSchema := map[string]*schema.Schema{
		"package_types": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
		"paths": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},

		"mime_types": {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
		"property": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"key": {
						Type:     schema.TypeString,
						Required: true,
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
		},
	}

	// A single repo schema is the same as all repositories, except it has name and BinMgrID fields
	singleRepoSchema := map[string]*schema.Schema{}
	singleRepoSchema["name"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	}
	singleRepoSchema["bin_mgr_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Default:  "default",
		Optional: true,
	}
	for key, value := range repoSchema {
		singleRepoSchema[key] = value
	}

	// An all repo schema has a collection of names
	allRepoSchema := map[string]*schema.Schema{}
	allRepoSchema["names"] = &schema.Schema{
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Optional: true,
	}
	for key, value := range repoSchema {
		allRepoSchema[key] = value
	}

	return &schema.Resource{
		Create: resourceWatchCreate,
		Read:   resourceWatchRead,
		Update: resourceWatchUpdate,
		Delete: resourceWatchDelete,
		Exists: resourceWatchExists,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"all_repositories": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: allRepoSchema,
				},
				MaxItems:      1,
				ConflictsWith: []string{"repository"},
			},
			"repository": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: singleRepoSchema,
				},
				ConflictsWith: []string{"all_repositories"},
			},
			"repository_paths": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_patterns": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"exclude_patterns": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
					},
				},
			},
			"all_builds": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bin_mgr_id": {
							Type:     schema.TypeString,
							Default:  "default",
							Optional: true,
						},
						"include_patterns": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"exclude_patterns": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
					},
				},
				MaxItems:      1,
				ConflictsWith: []string{"build"},
			},
			"build": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bin_mgr_id": {
							Type:     schema.TypeString,
							Default:  "default",
							Optional: true,
						},
					},
				},
				ConflictsWith: []string{"all_builds"},
			},
			"policy": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceWatchCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).XrayClient

	if client == nil {
		return fmt.Errorf("you must specify the xray_url in the provider to manage a watch")
	}

	params, err := unpackWatch(d, m)
	if err != nil {
		return err
	}

	_, err = client.CreateWatch(*params)
	if err != nil {
		return err
	}

	d.SetId(params.Name)

	return nil
}

func resourceWatchRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).XrayClient

	if client == nil {
		return fmt.Errorf("you must specify the xray_url in the provider to manage a watch")
	}

	watch, resp, err := client.GetWatch(d.Id())

	if resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		return nil
	} else if err != nil {
		return err
	}

	return packWatch(watch, d)
}

func resourceWatchUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).XrayClient

	if client == nil {
		return fmt.Errorf("you must specify the xray_url in the provider to manage a watch")
	}

	params, err := unpackWatch(d, m)
	if err != nil {
		return err
	}

	_, err = client.UpdateWatch(*params)
	if err != nil {
		return err
	}

	d.SetId(params.Name)
	return resourceWatchRead(d, m)
}

func resourceWatchDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*ArtClient).XrayClient

	if client == nil {
		return fmt.Errorf("you must specify the xray_url in the provider to manage a watch")
	}

	name := d.Get("name").(string)

	resp, err := client.DeleteWatch(name)
	if err != nil && resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return err
}

func resourceWatchExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*ArtClient).XrayClient

	if client == nil {
		return false, fmt.Errorf("you must specify the xray_url in the provider to manage a watch")
	}

	watchName := d.Id()
	_, resp, err := client.GetWatch(watchName)

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return true, err
}

func unpackWatch(d *schema.ResourceData, m interface{}) (*xrayutils.WatchParams, error) {
	artClient := m.(*ArtClient).ArtOld
	params := xrayutils.NewWatchParams()

	name := d.Get("name").(string)

	params.Name = name
	params.Description = d.Get("description").(string)
	params.Active = d.Get("active").(bool)

	policiesRaw := d.Get("policy").([]interface{})
	allRepositoriesRaw := d.Get("all_repositories").([]interface{})
	repositoriesRaw := d.Get("repository").([]interface{})
	allBuildsRaw := d.Get("all_builds").([]interface{})
	buildsRaw := d.Get("build").([]interface{})

	noPolicies := len(policiesRaw) == 0
	noResourcesToWatch := len(allRepositoriesRaw) == 0 && len(repositoriesRaw) == 0 && len(allBuildsRaw) == 0 && len(buildsRaw) == 0
	if (noResourcesToWatch || noPolicies) && params.Active {
		return nil, fmt.Errorf("`active` can only be true when at least 1 `policy` is defined and at least one of `all_repositories`, `repository`, `all_builds`, or `build` is defined")
	}

	err := unpackPolicies(policiesRaw, &params)
	if err != nil {
		return nil, err
	}

	unpackRepositoryPaths(d.Get("repository_paths"), &params)
	unpackAllRepositories(allRepositoriesRaw, &params)

	err = unpackRepository(repositoriesRaw, artClient, &params)
	if err != nil {
		return nil, err
	}

	unpackAllBuilds(allBuildsRaw, &params)
	unpackBuild(buildsRaw, &params)

	return &params, nil
}

func unpackPolicies(policiesRaw []interface{}, params *xrayutils.WatchParams) error {
	if len(policiesRaw) > 0 {
		params.Policies = make([]xrayutils.AssignedPolicy, 0)

		for _, policyRaw := range policiesRaw {
			policy := policyRaw.(map[string]interface{})
			assignedPolicy := xrayutils.AssignedPolicy{}
			assignedPolicy.Name = policy["name"].(string)
			assignedPolicy.Type = policy["type"].(string)

			if strings.ToLower(assignedPolicy.Type) != "security" && strings.ToLower(assignedPolicy.Type) != "license" {
				return fmt.Errorf("policy type %s must be security or license", assignedPolicy.Type)
			}

			params.Policies = append(params.Policies, assignedPolicy)
		}
	}

	return nil
}

func unpackRepositoryPaths(repositoryPaths interface{}, params *xrayutils.WatchParams) {
	if repositoryPaths == nil {
		return
	}

	repositoryPathsRaw := repositoryPaths.(*schema.Set).List()
	if len(repositoryPathsRaw) == 1 {
		repositoryPaths := repositoryPathsRaw[0].(map[string]interface{})
		includePatterns := castToStringArr(repositoryPaths["include_patterns"].(*schema.Set).List())
		sort.Strings(includePatterns)
		excludePatterns := castToStringArr(repositoryPaths["exclude_patterns"].(*schema.Set).List())
		sort.Strings(excludePatterns)

		params.Repositories.IncludePatterns = includePatterns
		params.Repositories.ExcludePatterns = excludePatterns
	}
}

func unpackAllRepositories(allRepositoriesRaw []interface{}, params *xrayutils.WatchParams) {
	if len(allRepositoriesRaw) == 1 {
		params.Repositories.Type = xrayutils.WatchRepositoriesAll

		if allRepositoriesRaw[0] != nil {
			allRepositories := allRepositoriesRaw[0].(map[string]interface{})
			names := castToStringArr(allRepositories["names"].(*schema.Set).List())
			sort.Strings(names)

			packageTypes := castToStringArr(allRepositories["package_types"].(*schema.Set).List())
			sort.Strings(packageTypes)
			paths := castToStringArr(allRepositories["paths"].(*schema.Set).List())
			sort.Strings(paths)
			mimeTypes := castToStringArr(allRepositories["mime_types"].(*schema.Set).List())
			sort.Strings(mimeTypes)

			params.Repositories.All.Filters.Names = names
			params.Repositories.All.Filters.PackageTypes = packageTypes
			params.Repositories.All.Filters.Paths = paths
			params.Repositories.All.Filters.MimeTypes = mimeTypes

			params.Repositories.All.Filters.Properties = map[string]string{}
			// Properties are nested in another level
			for _, rawValue := range allRepositories["property"].(*schema.Set).List() {
				value := rawValue.(map[string]interface{})
				params.Repositories.All.Filters.Properties[value["key"].(string)] = value["value"].(string)
			}
		}
	}
}

func unpackRepository(repositoriesRaw []interface{}, artClient *artifactoryold.Artifactory, params *xrayutils.WatchParams) error {
	if len(repositoriesRaw) > 0 {
		params.Repositories.Type = xrayutils.WatchRepositoriesByName
		// There can be 1 to M repositories
		params.Repositories.Repositories = make(map[string]xrayutils.WatchRepository, 0)

		for _, repositoryRaw := range repositoriesRaw {

			watchRepo := xrayutils.WatchRepository{}
			repository := repositoryRaw.(map[string]interface{})
			name := repository["name"].(string)

			// GetLocal will retrieve any repo type by name.
			repo, _, err := artClient.V1.Repositories.GetLocal(context.Background(), name)
			if err != nil {
				return err
			}

			// A watch requires the repo to have xray indexing on.
			if repo != nil {
				if *repo.XrayIndex == false {
					return fmt.Errorf("repo %s must have xray indexing on", name)
				}
			}

			binMgrID := repository["bin_mgr_id"].(string)
			watchRepo.Name = name
			watchRepo.BinMgrID = binMgrID

			packageTypes := castToStringArr(repository["package_types"].(*schema.Set).List())
			sort.Strings(packageTypes)
			paths := castToStringArr(repository["paths"].(*schema.Set).List())
			sort.Strings(paths)
			mimeTypes := castToStringArr(repository["mime_types"].(*schema.Set).List())
			sort.Strings(mimeTypes)

			watchRepo.Filters.PackageTypes = packageTypes
			watchRepo.Filters.Paths = paths
			watchRepo.Filters.MimeTypes = mimeTypes

			watchRepo.Filters.Properties = make(map[string]string)
			// Properties are nested in another level
			for _, rawValue := range repository["property"].(*schema.Set).List() {
				value := rawValue.(map[string]interface{})
				watchRepo.Filters.Properties[value["key"].(string)] = value["value"].(string)
			}

			params.Repositories.Repositories[name] = watchRepo
		}
	}

	return nil
}

func unpackAllBuilds(allBuildsRaw []interface{}, params *xrayutils.WatchParams) {
	if len(allBuildsRaw) == 1 {
		params.Builds.Type = xrayutils.WatchBuildAll
		allBuilds := allBuildsRaw[0].(map[string]interface{})
		params.Builds.All.BinMgrID = allBuilds["bin_mgr_id"].(string)

		includePatterns := castToStringArr(allBuilds["include_patterns"].(*schema.Set).List())
		sort.Strings(includePatterns)

		params.Builds.All.IncludePatterns = includePatterns

		excludePatterns := castToStringArr(allBuilds["exclude_patterns"].(*schema.Set).List())
		sort.Strings(excludePatterns)

		params.Builds.All.ExcludePatterns = excludePatterns
	}
}

func unpackBuild(buildsRaw []interface{}, params *xrayutils.WatchParams) {
	if len(buildsRaw) > 0 {
		params.Builds.Type = xrayutils.WatchBuildByName
		params.Builds.ByNames = make(map[string]xrayutils.WatchBuildsByNameParams, 0)

		for _, buildRaw := range buildsRaw {
			build := buildRaw.(map[string]interface{})
			buildParams := xrayutils.WatchBuildsByNameParams{}

			name := build["name"].(string)
			binMgrID := build["bin_mgr_id"].(string)

			buildParams.Name = name
			buildParams.BinMgrID = binMgrID

			params.Builds.ByNames[name] = buildParams
		}
	}
}

func packWatch(watch *xrayutils.WatchParams, d *schema.ResourceData) error {

	hasErr := false
	logErrors := cascadingErr(&hasErr)

	logErrors(d.Set("name", watch.Name))
	logErrors(d.Set("description", watch.Description))
	logErrors(d.Set("active", watch.Active))

	policies := packPolicies(watch.Policies)

	logErrors(d.Set("policy", policies))

	repositoryPaths := packPatternFilters(watch.Repositories.IncludePatterns, watch.Repositories.ExcludePatterns)

	if len(repositoryPaths) > 0 {
		logErrors(d.Set("repository_paths", []interface{}{repositoryPaths}))
	}

	switch watch.Repositories.Type {

	case xrayutils.WatchRepositoriesAll:
		data := make(map[string][]string)
		data["mime_types"] = watch.Repositories.All.Filters.MimeTypes
		data["names"] = watch.Repositories.All.Filters.Names
		data["package_types"] = watch.Repositories.All.Filters.PackageTypes
		data["paths"] = watch.Repositories.All.Filters.Paths
		allRepositories := packFilters(data, watch.Repositories.All.Filters.Properties)

		logErrors(d.Set("all_repositories", []interface{}{allRepositories}))

	case xrayutils.WatchRepositoriesByName:
		repositories := make([]interface{}, 0)

		for _, repo := range watch.Repositories.Repositories {
			data := make(map[string][]string)
			data["mime_types"] = repo.Filters.MimeTypes
			data["package_types"] = repo.Filters.PackageTypes
			data["paths"] = repo.Filters.Paths
			packedRepo := packFilters(data, repo.Filters.Properties)

			packedRepo["name"] = repo.Name
			packedRepo["bin_mgr_id"] = repo.BinMgrID

			repositories = append(repositories, packedRepo)
		}

		logErrors(d.Set("repository", repositories))
	}

	switch watch.Builds.Type {

	case xrayutils.WatchBuildAll:
		allBuilds := packPatternFilters(watch.Builds.All.IncludePatterns, watch.Builds.All.ExcludePatterns)

		allBuilds["bin_mgr_id"] = watch.Builds.All.BinMgrID

		logErrors(d.Set("all_builds", []interface{}{allBuilds}))

	case xrayutils.WatchBuildByName:
		builds := make([]interface{}, 0)

		for _, build := range watch.Builds.ByNames {
			packedBuild := make(map[string]interface{})
			packedBuild["name"] = build.Name
			packedBuild["bin_mgr_id"] = build.BinMgrID

			builds = append(builds, packedBuild)
		}

		logErrors(d.Set("build", builds))
	}

	if hasErr {
		return fmt.Errorf("failed to marshal watch")
	}

	return nil
}

func packPolicies(policies []xrayutils.AssignedPolicy) []interface{} {
	packedPolicies := make([]interface{}, 0)

	for _, policy := range policies {
		packedPolicy := make(map[string]interface{})

		packedPolicy["name"] = policy.Name
		packedPolicy["type"] = policy.Type

		packedPolicies = append(packedPolicies, packedPolicy)
	}

	return packedPolicies
}

func packFilters(data map[string][]string, properties map[string]string) map[string]interface{} {
	packedFilters := make(map[string]interface{})

	for k, v := range data {
		if len(v) > 0 {
			packedFilters[k] = schema.NewSet(schema.HashString, castToInterfaceArr(v))
		} else {
			packedFilters[k] = schema.NewSet(schema.HashString, []interface{}{})
		}
	}

	propertyList := make([]interface{}, 0)

	for propKey, propVal := range properties {
		property := map[string]string{}
		property["key"] = propKey
		property["value"] = propVal
		propertyList = append(propertyList, property)
	}
	packedFilters["property"] = propertyList

	return packedFilters
}

func packPatternFilters(includePatterns []string, excludePatterns []string) map[string]interface{} {
	packedPatternFilters := make(map[string]interface{})

	if len(includePatterns) > 0 {
		packedPatternFilters["include_patterns"] = schema.NewSet(schema.HashString, castToInterfaceArr(includePatterns))
	}

	if len(excludePatterns) > 0 {
		packedPatternFilters["exclude_patterns"] = schema.NewSet(schema.HashString, castToInterfaceArr(excludePatterns))
	}

	return packedPatternFilters
}
