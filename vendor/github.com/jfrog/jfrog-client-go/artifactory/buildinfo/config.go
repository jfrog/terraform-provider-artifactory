package buildinfo

import (
	"path/filepath"
	"strings"

	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const BuildInfoEnvPrefix = "buildInfo.env."

type Configuration struct {
	ArtDetails auth.ServiceDetails
	BuildUrl   string
	DryRun     bool
	EnvInclude string
	EnvExclude string
}

func (config *Configuration) GetArtifactoryDetails() auth.ServiceDetails {
	return config.ArtDetails
}

func (config *Configuration) SetArtifactoryDetails(artDetails auth.ServiceDetails) {
	config.ArtDetails = artDetails
}

func (config *Configuration) IsDryRun() bool {
	return config.DryRun
}

type Filter func(map[string]string) (map[string]string, error)

// IncludeFilter returns a function used to filter entries of a map based on key
func (config Configuration) IncludeFilter() Filter {
	pats := strings.Split(config.EnvInclude, ";")
	return func(tempMap map[string]string) (map[string]string, error) {
		result := make(map[string]string)
		for k, v := range tempMap {
			for _, filterPattern := range pats {
				matched, err := filepath.Match(strings.ToLower(filterPattern), strings.ToLower(strings.TrimPrefix(k, BuildInfoEnvPrefix)))
				if errorutils.CheckError(err) != nil {
					return nil, err
				}
				if matched {
					result[k] = v
					break
				}
			}
		}
		return result, nil
	}
}

// ExcludeFilter returns a function used to filter entries of a map based on key
func (config Configuration) ExcludeFilter() Filter {
	pats := strings.Split(config.EnvExclude, ";")
	return func(tempMap map[string]string) (map[string]string, error) {
		result := make(map[string]string)
		for k, v := range tempMap {
			include := true
			for _, filterPattern := range pats {
				matched, err := filepath.Match(strings.ToLower(filterPattern), strings.ToLower(strings.TrimPrefix(k, BuildInfoEnvPrefix)))
				if errorutils.CheckError(err) != nil {
					return nil, err
				}
				if matched {
					include = false
					break
				}
			}
			if include {
				result[k] = v
			}
		}
		return result, nil
	}
}
