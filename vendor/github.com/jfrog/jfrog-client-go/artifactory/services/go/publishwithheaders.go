package _go

import (
	"net/http"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/version"
)

func init() {
	register(&publishWithHeader{})
}

// Support for Artifactory older then 6.5.0 API
type publishWithHeader struct {
}

func (pwh *publishWithHeader) isCompatible(artifactoryVersion string) bool {
	propertiesApi := "6.5.0"
	version := version.NewVersion(artifactoryVersion)
	if version.Compare(propertiesApi) > 0 {
		return true
	}
	return false
}

func (pwh *publishWithHeader) PublishPackage(params GoParams, client *rthttpclient.ArtifactoryHttpClient, ArtDetails auth.ServiceDetails) error {
	url, err := utils.BuildArtifactoryUrl(ArtDetails.GetUrl(), "api/go/"+params.GetTargetRepo(), make(map[string]string))
	clientDetails := ArtDetails.CreateHttpClientDetails()
	addHeaders(params, &clientDetails)
	err = addPropertiesHeaders(params.GetProps(), &clientDetails.Headers)
	if err != nil {
		return err
	}
	resp, _, err := client.UploadFile(params.GetZipPath(), url, "", &clientDetails, GoUploadRetries, nil)
	if err != nil {
		return err
	}
	return errorutils.CheckResponseStatus(resp, http.StatusCreated)
}

func addPropertiesHeaders(props string, headers *map[string]string) error {
	properties, err := utils.ParseProperties(props, utils.JoinCommas)
	if err != nil {
		return err
	}
	headersMap := properties.ToHeadersMap()
	for k, v := range headersMap {
		utils.AddHeader("X-ARTIFACTORY-PROPERTY-"+k, v, headers)
	}
	return nil
}
