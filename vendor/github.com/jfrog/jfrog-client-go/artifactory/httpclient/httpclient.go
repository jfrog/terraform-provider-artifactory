package httpclient

import (
	"io"
	"net/http"
	"net/url"

	"github.com/jfrog/jfrog-client-go/auth"
	"github.com/jfrog/jfrog-client-go/httpclient"
	ioutils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
)

type ArtifactoryHttpClient struct {
	httpClient *httpclient.HttpClient
	ArtDetails *auth.ServiceDetails
}

func (rtc *ArtifactoryHttpClient) SendGet(url string, followRedirect bool, httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.SendGet(url, followRedirect, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) SendPost(url string, content []byte, httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.SendPost(url, content, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) SendPostLeaveBodyOpen(url string, content []byte, httpClientsDetails *httputils.HttpClientDetails) (*http.Response, error) {
	if err := (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails); err != nil {
		return nil, err
	}
	return rtc.httpClient.SendPostLeaveBodyOpen(url, content, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) SendPostForm(url string, data url.Values, httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	httpClientsDetails.Headers["Content-Type"] = "application/x-www-form-urlencoded"
	return rtc.SendPost(url, []byte(data.Encode()), httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) SendPatch(url string, content []byte, httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.SendPatch(url, content, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) SendDelete(url string, content []byte, httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.SendDelete(url, content, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) SendHead(url string, httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.SendHead(url, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) SendPut(url string, content []byte, httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.SendPut(url, content, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) Send(method string, url string, content []byte, followRedirect bool, closeBody bool,
	httpClientsDetails *httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.Send(method, url, content, followRedirect, closeBody, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) UploadFile(localPath, url, logMsgPrefix string,
	httpClientsDetails *httputils.HttpClientDetails, retries int, progress ioutils.Progress) (resp *http.Response, body []byte, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.UploadFile(localPath, url, logMsgPrefix, *httpClientsDetails, retries, progress)
}

func (rtc *ArtifactoryHttpClient) ReadRemoteFile(downloadPath string, httpClientsDetails *httputils.HttpClientDetails) (ioReaderCloser io.ReadCloser, resp *http.Response, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.ReadRemoteFile(downloadPath, *httpClientsDetails)
}

func (rtc *ArtifactoryHttpClient) DownloadFileWithProgress(downloadFileDetails *httpclient.DownloadFileDetails, logMsgPrefix string,
	httpClientsDetails *httputils.HttpClientDetails, retries int, isExplode bool, progress ioutils.Progress) (resp *http.Response, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.DownloadFileWithProgress(downloadFileDetails, logMsgPrefix, *httpClientsDetails, retries, isExplode, progress)
}

func (rtc *ArtifactoryHttpClient) DownloadFile(downloadFileDetails *httpclient.DownloadFileDetails, logMsgPrefix string,
	httpClientsDetails *httputils.HttpClientDetails, retries int, isExplode bool) (resp *http.Response, err error) {
	return rtc.DownloadFileWithProgress(downloadFileDetails, logMsgPrefix, httpClientsDetails, retries, isExplode, nil)
}

func (rtc *ArtifactoryHttpClient) DownloadFileConcurrently(flags httpclient.ConcurrentDownloadFlags,
	logMsgPrefix string, httpClientsDetails *httputils.HttpClientDetails, progress ioutils.Progress) (resp *http.Response, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.DownloadFileConcurrently(flags, logMsgPrefix, *httpClientsDetails, progress)
}

func (rtc *ArtifactoryHttpClient) IsAcceptRanges(downloadUrl string, httpClientsDetails *httputils.HttpClientDetails) (isAcceptRanges bool, resp *http.Response, err error) {
	err = (*rtc.ArtDetails).RunPreRequestInterceptors(httpClientsDetails)
	if err != nil {
		return
	}
	return rtc.httpClient.IsAcceptRanges(downloadUrl, *httpClientsDetails)
}
