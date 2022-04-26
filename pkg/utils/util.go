package utils

import (
	"fmt"
)

type SupportedRepoClasses struct {
	RepoLayoutRef      string
	SupportedRepoTypes map[string]bool
}

//Consolidated list of Default Repo Layout for all Package Types with active Repo Types
var defaultRepoLayoutMap = map[string]SupportedRepoClasses{
	"alpine":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"bower":     {RepoLayoutRef: "bower-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cran":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cargo":     {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "federated": true}},
	"chef":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"cocoapods": {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "federated": true}},
	"composer":  {RepoLayoutRef: "composer-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"conan":     {RepoLayoutRef: "conan-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"conda":     {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"debian":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"docker":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gems":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"generic":   {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gitlfs":    {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"go":        {RepoLayoutRef: "go-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"gradle":    {RepoLayoutRef: "maven-2-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"helm":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"ivy":       {RepoLayoutRef: "ivy-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"maven":     {RepoLayoutRef: "maven-2-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"npm":       {RepoLayoutRef: "npm-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"nuget":     {RepoLayoutRef: "nuget-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"opkg":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"p2":        {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"remote": true, "virtual": true}},
	"pub":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"puppet":    {RepoLayoutRef: "puppet-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"pypi":      {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"sbt":       {RepoLayoutRef: "sbt-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
	"vagrant":   {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "federated": true}},
	"vcs":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"remote": true}},
	"rpm":       {RepoLayoutRef: "simple-default", SupportedRepoTypes: map[string]bool{"local": true, "remote": true, "virtual": true, "federated": true}},
}

//Return the default repo layout by Repository Type & Package Type
func GetDefaultRepoLayoutRef(repositoryType string, packageType string) func() (interface{}, error) {
	return func() (interface{}, error) {
		if v, ok := defaultRepoLayoutMap[packageType].SupportedRepoTypes[repositoryType]; ok && v {
			return defaultRepoLayoutMap[packageType].RepoLayoutRef, nil
		}
		return "", fmt.Errorf("default repo layout not found for repository type %v & package type %v", repositoryType, packageType)
	}
}
