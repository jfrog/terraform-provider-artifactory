package services

import (
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/jfrog/gofrog/parallel"
	rthttpclient "github.com/jfrog/jfrog-client-go/artifactory/httpclient"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/auth"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/content"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type PropsService struct {
	client     *rthttpclient.ArtifactoryHttpClient
	ArtDetails auth.ServiceDetails
	Threads    int
}

func NewPropsService(client *rthttpclient.ArtifactoryHttpClient) *PropsService {
	return &PropsService{client: client}
}

func (ps *PropsService) GetArtifactoryDetails() auth.ServiceDetails {
	return ps.ArtDetails
}

func (ps *PropsService) SetArtifactoryDetails(rt auth.ServiceDetails) {
	ps.ArtDetails = rt
}

func (ps *PropsService) IsDryRun() bool {
	return false
}

func (ps *PropsService) GetThreads() int {
	return ps.Threads
}

func (ps *PropsService) SetProps(propsParams PropsParams) (int, error) {
	log.Info("Setting properties...")
	totalSuccess, err := ps.performRequest(propsParams, false)
	if err == nil {
		log.Info("Done setting properties.")
	}
	return totalSuccess, err
}

func (ps *PropsService) DeleteProps(propsParams PropsParams) (int, error) {
	log.Info("Deleting properties...")
	totalSuccess, err := ps.performRequest(propsParams, true)
	if err == nil {
		log.Info("Done deleting properties.")
	}
	return totalSuccess, err
}

type PropsParams struct {
	Reader *content.ContentReader
	Props  string
}

func (sp *PropsParams) GetReader() *content.ContentReader {
	return sp.Reader
}

func (sp *PropsParams) GetProps() string {
	return sp.Props
}

func (ps *PropsService) performRequest(propsParams PropsParams, isDelete bool) (int, error) {
	var encodedParam string
	if !isDelete {
		props, err := utils.ParseProperties(propsParams.GetProps(), utils.JoinCommas)
		if err != nil {
			return 0, err
		}
		encodedParam = props.ToEncodedString()
	} else {
		propList := strings.Split(propsParams.GetProps(), ",")
		for _, prop := range propList {
			encodedParam += url.QueryEscape(prop) + ","
		}
		// Remove trailing comma
		if strings.HasSuffix(encodedParam, ",") {
			encodedParam = encodedParam[:len(encodedParam)-1]
		}

	}

	successCounters := make([]int, ps.GetThreads())
	producerConsumer := parallel.NewBounedRunner(ps.GetThreads(), false)
	errorsQueue := clientutils.NewErrorsQueue(1)
	reader := propsParams.GetReader()
	go func() {
		for resultItem := new(utils.ResultItem); reader.NextRecord(resultItem) == nil; resultItem = new(utils.ResultItem) {
			relativePath := resultItem.GetItemRelativePath()
			setPropsTask := func(threadId int) error {
				var err error
				logMsgPrefix := clientutils.GetLogMsgPrefix(threadId, ps.IsDryRun())

				restApi := path.Join("api", "storage", relativePath)
				setPropertiesUrl, err := utils.BuildArtifactoryUrl(ps.GetArtifactoryDetails().GetUrl(), restApi, make(map[string]string))
				if err != nil {
					return err
				}
				setPropertiesUrl += "?properties=" + encodedParam

				var resp *http.Response
				if isDelete {
					resp, _, err = ps.sendDeleteRequest(logMsgPrefix, relativePath, setPropertiesUrl)
				} else {
					resp, _, err = ps.sendPutRequest(logMsgPrefix, relativePath, setPropertiesUrl)
				}

				if err != nil {
					return err
				}
				if err = errorutils.CheckResponseStatus(resp, http.StatusNoContent); err != nil {
					return errorutils.CheckError(err)
				}
				successCounters[threadId]++
				return nil
			}

			producerConsumer.AddTaskWithError(setPropsTask, errorsQueue.AddError)
		}
		defer producerConsumer.Done()
		if err := reader.GetError(); err != nil {
			errorsQueue.AddError(err)
		}
		reader.Reset()
	}()

	producerConsumer.Run()
	totalSuccess := 0
	for _, v := range successCounters {
		totalSuccess += v
	}
	return totalSuccess, errorsQueue.GetError()
}

func (ps *PropsService) sendDeleteRequest(logMsgPrefix, relativePath, setPropertiesUrl string) (resp *http.Response, body []byte, err error) {
	log.Info(logMsgPrefix+"Deleting properties on:", relativePath)
	log.Debug(logMsgPrefix+"Sending delete properties request:", setPropertiesUrl)
	httpClientsDetails := ps.GetArtifactoryDetails().CreateHttpClientDetails()
	resp, body, err = ps.client.SendDelete(setPropertiesUrl, nil, &httpClientsDetails)
	return
}

func (ps *PropsService) sendPutRequest(logMsgPrefix, relativePath, setPropertiesUrl string) (resp *http.Response, body []byte, err error) {
	log.Info(logMsgPrefix+"Setting properties on:", relativePath)
	log.Debug(logMsgPrefix+"Sending set properties request:", setPropertiesUrl)
	httpClientsDetails := ps.GetArtifactoryDetails().CreateHttpClientDetails()
	resp, body, err = ps.client.SendPut(setPropertiesUrl, nil, &httpClientsDetails)
	return
}

func NewPropsParams() PropsParams {
	return PropsParams{}
}
