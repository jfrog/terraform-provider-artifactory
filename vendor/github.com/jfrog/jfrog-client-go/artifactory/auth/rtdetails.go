package auth

import (
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

func NewArtifactoryDetails() auth.ServiceDetails {
	return &artifactoryDetails{}
}

type artifactoryDetails struct {
	auth.CommonConfigFields
}

func (rt *artifactoryDetails) GetVersion() (string, error) {
	var err error
	if rt.Version == "" {
		rt.Version, err = rt.getArtifactoryVersion()
		if err != nil {
			return "", err
		}
		log.Debug("The Artifactory version is:", rt.Version)
	}
	return rt.Version, nil
}

func (rt *artifactoryDetails) getArtifactoryVersion() (string, error) {
	cd := auth.ServiceDetails(rt)
	serviceConfig, err := config.NewConfigBuilder().
		SetServiceDetails(cd).
		SetCertificatesPath(cd.GetClientCertPath()).
		Build()
	sm, err := artifactory.New(&cd, serviceConfig)
	if err != nil {
		return "", err
	}
	return sm.GetVersion()
}
