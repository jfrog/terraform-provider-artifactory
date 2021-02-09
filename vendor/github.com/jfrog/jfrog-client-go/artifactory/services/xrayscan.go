package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/httpclient"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const SCAN_BUILD_API_URL = "api/xray/scanBuild"
const XRAY_SCAN_RETRY_CONSECUTIVE_RETRIES = 10           // Retrying to resume the scan 10 times after a stable connection
const XRAY_SCAN_CONNECTION_TIMEOUT = 90 * time.Second    // Expecting \r\n every 30 seconds
const XRAY_SCAN_SLEEP_BETWEEN_RETRIES = 15 * time.Second // 15 seconds sleep between retry
const XRAY_SCAN_STABLE_CONNECTION_WINDOW = 100 * time.Second
const XRAY_FATAL_FAIL_STATUS = -1

type XrayScanService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
}

func NewXrayScanService(client *rthttpclient.ArtifactoryHttpClient) *XrayScanService {
	return &XrayScanService{client: client}
}

func (ps *XrayScanService) ScanBuild(scanParams XrayScanParams) ([]byte, error) {
	url := ps.ArtDetails.GetUrl()
	requestFullUrl, err := utils.BuildArtifactoryUrl(url, SCAN_BUILD_API_URL, make(map[string]string))
	if err != nil {
		return []byte{}, err
	}
	data := XrayScanBody{
		BuildName:   scanParams.GetBuildName(),
		BuildNumber: scanParams.GetBuildNumber(),
		Context:     clientutils.GetUserAgent(),
	}

	requestContent, err := json.Marshal(data)
	if err != nil {
		return []byte{}, errorutils.CheckError(err)
	}

	connection := httpclient.RetryableConnection{
		ReadTimeout:            XRAY_SCAN_CONNECTION_TIMEOUT,
		RetriesNum:             XRAY_SCAN_RETRY_CONSECUTIVE_RETRIES,
		StableConnectionWindow: XRAY_SCAN_STABLE_CONNECTION_WINDOW,
		SleepBetweenRetries:    XRAY_SCAN_SLEEP_BETWEEN_RETRIES,
		ConnectHandler: func() (*http.Response, error) {
			return ps.execScanRequest(requestFullUrl, requestContent)
		},
		ErrorHandler: func(content []byte) error {
			return checkForXrayResponseError(content, true)
		},
	}
	result, err := connection.Do()
	if err != nil {
		return []byte{}, err
	}

	return result, nil
}

func isFatalScanError(errResp *errorResponse) bool {
	if errResp == nil {
		return false
	}
	for _, v := range errResp.Errors {
		if v.Status == XRAY_FATAL_FAIL_STATUS {
			return true
		}
	}
	return false
}

func checkForXrayResponseError(content []byte, ignoreFatalError bool) error {
	respErrors := &errorResponse{}
	err := json.Unmarshal(content, respErrors)
	if errorutils.CheckError(err) != nil {
		return err
	}

	if respErrors.Errors == nil {
		return nil
	}

	if ignoreFatalError && isFatalScanError(respErrors) {
		// fatal error should be interpreted as no errors so no more retries will accrue
		return nil
	}
	return errorutils.CheckError(errors.New("Artifactory response: " + string(content)))
}

func (ps *XrayScanService) execScanRequest(url string, content []byte) (*http.Response, error) {
	httpClientsDetails := ps.ArtDetails.CreateHttpClientDetails()
	utils.SetContentType("application/json", &httpClientsDetails.Headers)

	// The scan build operation can take a long time to finish.
	// To keep the connection open, when Xray starts scanning the build, it starts sending new-lines
	// on the open channel. This tells the client that the operation is still in progress and the
	// connection does not get timed out.
	// We need make sure the new-lines are not buffered on the nginx and are flushed
	// as soon as Xray sends them.
	utils.DisableAccelBuffering(&httpClientsDetails.Headers)

	resp, _, _, err := ps.client.Send("POST", url, content, true, false, &httpClientsDetails)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		err = errorutils.CheckError(errors.New("Artifactory Response: " + resp.Status))
	}
	return resp, err
}

type errorResponse struct {
	Errors []errorsStatusResponse `json:"errors,omitempty"`
}

type errorsStatusResponse struct {
	Status int `json:"status,omitempty"`
}

type XrayScanBody struct {
	BuildName   string `json:"buildName,omitempty"`
	BuildNumber string `json:"buildNumber,omitempty"`
	Context     string `json:"context,omitempty"`
}

type XrayScanParams struct {
	BuildName   string
	BuildNumber string
}

func (bp *XrayScanParams) GetBuildName() string {
	return bp.BuildName
}

func (bp *XrayScanParams) GetBuildNumber() string {
	return bp.BuildNumber
}

func NewXrayScanParams() XrayScanParams {
	return XrayScanParams{}
}
