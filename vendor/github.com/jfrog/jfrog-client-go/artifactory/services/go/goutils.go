package _go

import (
	"encoding/base64"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"strings"
)

const GoUploadRetries = 3

func addHeaders(params GoParams, clientDetails *httputils.HttpClientDetails) {
	utils.AddHeader("X-GO-MODULE-VERSION", params.GetVersion(), &clientDetails.Headers)
	utils.AddHeader("X-GO-MODULE-CONTENT", base64.StdEncoding.EncodeToString(params.GetModContent()), &clientDetails.Headers)
}

func CreateUrlPath(moduleId, version, props, extension string, url *string) error {
	*url = strings.Join([]string{*url, moduleId, "@v", version + extension}, "/")
	properties, err := utils.ParseProperties(props, utils.JoinCommas)
	if err != nil {
		return err
	}

	*url = strings.Join([]string{*url, properties.ToEncodedString()}, ";")
	if strings.HasSuffix(*url, ";") {
		tempUrl := *url
		tempUrl = tempUrl[:len(tempUrl)-1]
		*url = tempUrl
	}
	return nil
}
