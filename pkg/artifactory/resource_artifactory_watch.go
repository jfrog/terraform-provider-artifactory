package artifactory

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

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
	if len(policiesRaw) > 0 {
		params.Policies = make([]xrayutils.AssignedPolicy, 0)

		for _, policyRaw := range policiesRaw {
			policy := policyRaw.(map[string]interface{})
			assignedPolicy := xrayutils.AssignedPolicy{}
			assignedPolicy.Name = policy["name"].(string)
			assignedPolicy.Type = policy["type"].(string)

			if strings.ToLower(assignedPolicy.Type) != "security" && strings.ToLower(assignedPolicy.Type) != "license" {
				return nil, fmt.Errorf("policy type %s must be security or license", assignedPolicy.Type)
			}

			params.Policies = append(params.Policies, assignedPolicy)
		}
	}

	if d.Get("repository_paths") != nil {
		repositoryPathsRaw := d.Get("repository_paths").(*schema.Set).List()
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

	allRepositoriesRaw := d.Get("all_repositories").([]interface{})
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

	repositoriesRaw := d.Get("repository").([]interface{})
	if len(repositoriesRaw) > 0 {
		params.Repositories.Type = xrayutils.WatchRepositoriesByName
		// There can be 1 to M repositories
		params.Repositories.Repositories = make(map[string]xrayutils.WatchRepository, 0)

		for _, repositoryRaw := range repositoriesRaw {

			watchRepo := xrayutils.WatchRepository{}
			repository := repositoryRaw.(map[string]interface{})
			name := repository["name"].(string)

			localRepo, _, _ := artClient.V1.Repositories.GetLocal(context.Background(), name)
			if localRepo != nil {
				if *localRepo.XrayIndex == false {
					return nil, fmt.Errorf("repo %s must have xray indexing on", name)
				}
			}

			remoteRepo, _, _ := artClient.V1.Repositories.GetRemote(context.Background(), name)
			if remoteRepo != nil {
				if *remoteRepo.XrayIndex == false {
					return nil, fmt.Errorf("repo %s must have xray indexing on", name)
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

	allBuildsRaw := d.Get("all_builds").([]interface{})
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

	buildsRaw := d.Get("build").([]interface{})
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

	noResourcesToWatch := len(buildsRaw) == 0 && len(repositoriesRaw) == 0 && len(allBuildsRaw) == 0 && len(allRepositoriesRaw) == 0
	noPolicies := len(policiesRaw) == 0
	if (noResourcesToWatch || noPolicies) && params.Active {
		return nil, fmt.Errorf("`active` can only be true when at least 1 `policy` is defined and at least one of `all_repositories`, `repository`, `all_builds`, or `build` is defined")
	}

	return &params, nil
}

func packWatch(watch *xrayutils.WatchParams, d *schema.ResourceData) error {

	hasErr := false
	logErrors := cascadingErr(&hasErr)

	logErrors(d.Set("name", watch.Name))
	logErrors(d.Set("description", watch.Description))
	logErrors(d.Set("active", watch.Active))

	policies := make([]interface{}, 0)

	for _, policy := range watch.Policies {
		packedPolicy := make(map[string]interface{})

		packedPolicy["name"] = policy.Name
		packedPolicy["type"] = policy.Type

		policies = append(policies, packedPolicy)
	}

	logErrors(d.Set("policy", policies))

	repositoryPaths := make(map[string]interface{})
	if len(watch.Repositories.IncludePatterns) > 0 {
		repositoryPaths["include_patterns"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Repositories.IncludePatterns))
	}

	if len(watch.Repositories.ExcludePatterns) > 0 {
		repositoryPaths["exclude_patterns"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Repositories.ExcludePatterns))
	}

	if len(repositoryPaths) > 0 {
		logErrors(d.Set("repository_paths", []interface{}{repositoryPaths}))
	}

	switch watch.Repositories.Type {

	case xrayutils.WatchRepositoriesAll:
		allRepositories := make(map[string]interface{})
		if len(watch.Repositories.All.Filters.MimeTypes) > 0 {
			allRepositories["mime_types"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Repositories.All.Filters.MimeTypes))
		} else {
			allRepositories["mime_types"] = schema.NewSet(schema.HashString, []interface{}{})
		}

		if len(watch.Repositories.All.Filters.Names) > 0 {
			allRepositories["names"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Repositories.All.Filters.Names))
		} else {
			allRepositories["names"] = schema.NewSet(schema.HashString, []interface{}{})
		}
		if len(watch.Repositories.All.Filters.PackageTypes) > 0 {
			allRepositories["package_types"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Repositories.All.Filters.PackageTypes))
		} else {
			allRepositories["package_types"] = schema.NewSet(schema.HashString, []interface{}{})
		}
		if len(watch.Repositories.All.Filters.Paths) > 0 {
			allRepositories["paths"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Repositories.All.Filters.Paths))
		} else {
			allRepositories["paths"] = schema.NewSet(schema.HashString, []interface{}{})
		}
		if len(watch.Repositories.All.Filters.Properties) > 0 {
			properties := make([]interface{}, 0)

			for propKey, propVal := range watch.Repositories.All.Filters.Properties {
				property := map[string]string{}
				property["key"] = propKey
				property["value"] = propVal
				properties = append(properties, property)
			}
			allRepositories["property"] = properties
		}

		logErrors(d.Set("all_repositories", []interface{}{allRepositories}))

	case xrayutils.WatchRepositoriesByName:
		repositories := make([]interface{}, 0)

		for _, repo := range watch.Repositories.Repositories {
			packedRepo := make(map[string]interface{})

			packedRepo["name"] = repo.Name
			packedRepo["bin_mgr_id"] = repo.BinMgrID

			if len(repo.Filters.MimeTypes) > 0 {
				packedRepo["mime_types"] = schema.NewSet(schema.HashString, castToInterfaceArr(repo.Filters.MimeTypes))
			}
			if len(repo.Filters.PackageTypes) > 0 {
				packedRepo["package_types"] = schema.NewSet(schema.HashString, castToInterfaceArr(repo.Filters.PackageTypes))
			}
			if len(repo.Filters.Paths) > 0 {
				packedRepo["paths"] = schema.NewSet(schema.HashString, castToInterfaceArr(repo.Filters.Paths))
			}
			if len(repo.Filters.Properties) > 0 {
				properties := make([]interface{}, 0)

				for propKey, propVal := range repo.Filters.Properties {
					property := map[string]string{}
					property["key"] = propKey
					property["value"] = propVal
					properties = append(properties, property)
				}
				packedRepo["property"] = properties
			}

			repositories = append(repositories, packedRepo)
		}

		logErrors(d.Set("repository", repositories))
	}

	switch watch.Builds.Type {

	case xrayutils.WatchBuildAll:
		allBuilds := make(map[string]interface{})

		allBuilds["bin_mgr_id"] = watch.Builds.All.BinMgrID

		if len(watch.Builds.All.IncludePatterns) > 0 {
			allBuilds["include_patterns"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Builds.All.IncludePatterns))
		}

		if len(watch.Builds.All.ExcludePatterns) > 0 {
			allBuilds["exclude_patterns"] = schema.NewSet(schema.HashString, castToInterfaceArr(watch.Builds.All.ExcludePatterns))
		}

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
