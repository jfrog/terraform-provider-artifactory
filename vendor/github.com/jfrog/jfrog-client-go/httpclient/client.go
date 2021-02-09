package httpclient

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	ioutils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/mholt/archiver"
)

func (jc *HttpClient) sendGetLeaveBodyOpen(url string, followRedirect bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	return jc.Send("GET", url, nil, followRedirect, false, httpClientsDetails)
}

func (jc *HttpClient) SendPostLeaveBodyOpen(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, err error) {
	resp, _, _, err = jc.Send("POST", url, content, true, false, httpClientsDetails)
	return
}

type HttpClient struct {
	Client *http.Client
}

func (jc *HttpClient) sendGetForFileDownload(url string, followRedirect bool, httpClientsDetails httputils.HttpClientDetails, currentSplit int) (resp *http.Response, redirectUrl string, err error) {
	resp, _, redirectUrl, err = jc.sendGetLeaveBodyOpen(url, followRedirect, httpClientsDetails)
	return
}

func (jc *HttpClient) Stream(url string, httpClientsDetails httputils.HttpClientDetails) (*http.Response, []byte, string, error) {
	return jc.sendGetLeaveBodyOpen(url, true, httpClientsDetails)
}

func (jc *HttpClient) SendGet(url string, followRedirect bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	return jc.Send("GET", url, nil, followRedirect, true, httpClientsDetails)
}

func (jc *HttpClient) SendPost(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("POST", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendPatch(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("PATCH", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendDelete(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("DELETE", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendHead(url string, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("HEAD", url, nil, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) SendPut(url string, content []byte, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, body []byte, err error) {
	resp, body, _, err = jc.Send("PUT", url, content, true, true, httpClientsDetails)
	return
}

func (jc *HttpClient) Send(method, url string, content []byte, followRedirect, closeBody bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	var req *http.Request
	log.Debug(fmt.Sprintf("Sending HTTP %s request to: %s", method, url))
	if content != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if errorutils.CheckError(err) != nil {
		return nil, nil, "", err
	}

	return jc.doRequest(req, content, followRedirect, closeBody, httpClientsDetails)
}

func (jc *HttpClient) doRequest(req *http.Request, content []byte, followRedirect bool, closeBody bool, httpClientsDetails httputils.HttpClientDetails) (resp *http.Response, respBody []byte, redirectUrl string, err error) {
	req.Close = true
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)
	copyHeaders(httpClientsDetails, req)

	if !followRedirect || (followRedirect && req.Method == "POST") {
		jc.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			redirectUrl = req.URL.String()
			return errors.New("redirect")
		}
	}

	resp, err = jc.Client.Do(req)
	jc.Client.CheckRedirect = nil

	if err != nil && redirectUrl != "" {
		if !followRedirect {
			log.Debug("Blocking HTTP redirect to ", redirectUrl)
			return
		}
		// Due to security reasons, there's no built in HTTP redirect in the HTTP Client
		// for POST requests. We therefore implement the redirect on our own.
		if req.Method == "POST" {
			log.Debug("HTTP redirecting to ", redirectUrl)
			resp, respBody, err = jc.SendPost(redirectUrl, content, httpClientsDetails)
			redirectUrl = ""
			return
		}
	}

	err = errorutils.CheckError(err)
	if err != nil {
		return
	}
	if closeBody {
		defer resp.Body.Close()
		respBody, _ = ioutil.ReadAll(resp.Body)
	}
	return
}

func copyHeaders(httpClientsDetails httputils.HttpClientDetails, req *http.Request) {
	if httpClientsDetails.Headers != nil {
		for name := range httpClientsDetails.Headers {
			req.Header.Set(name, httpClientsDetails.Headers[name])
		}
	}
}

func setRequestHeaders(httpClientsDetails httputils.HttpClientDetails, size int64, req *http.Request) {
	copyHeaders(httpClientsDetails, req)
	length := strconv.FormatInt(size, 10)
	req.Header.Set("Content-Length", length)
}

// You may implement the log.Progress interface, or pass nil to run without progress display.
func (jc *HttpClient) UploadFile(localPath, url, logMsgPrefix string, httpClientsDetails httputils.HttpClientDetails,
	retries int, progress ioutils.Progress) (resp *http.Response, body []byte, err error) {
	retryExecutor := utils.RetryExecutor{
		MaxRetries:      retries,
		RetriesInterval: 0,
		ErrorMessage:    fmt.Sprintf("Failure occurred while uploading to %s", url),
		LogMsgPrefix:    logMsgPrefix,
		ExecutionHandler: func() (bool, error) {
			resp, body, err = jc.doUploadFile(localPath, url, httpClientsDetails, progress)
			if err != nil {
				return true, err
			}
			// Response must not be nil
			if resp == nil {
				return false, errorutils.CheckError(errors.New(fmt.Sprintf("%sReceived empty response from file upload", logMsgPrefix)))
			}
			// If response-code < 500, should not retry
			if resp.StatusCode < 500 {
				return false, nil
			}
			// Perform retry
			log.Warn(fmt.Sprintf("%sArtifactory response: %s", logMsgPrefix, resp.Status))
			return true, nil
		},
	}

	err = retryExecutor.Execute()
	return
}

func (jc *HttpClient) doUploadFile(localPath, url string, httpClientsDetails httputils.HttpClientDetails, progress ioutils.Progress) (*http.Response, []byte, error) {
	var file *os.File
	var err error
	if localPath != "" {
		file, err = os.Open(localPath)
		defer file.Close()
		if errorutils.CheckError(err) != nil {
			return nil, nil, err
		}
	}

	size, err := fileutils.GetFileSize(file)
	if err != nil {
		return nil, nil, err
	}

	reqContent := fileutils.GetUploadRequestContent(file)
	var reader io.Reader
	if file != nil && progress != nil {
		progressId := progress.New(size, "Uploading", localPath)
		reader = progress.ReadWithProgress(progressId, reqContent)
		defer progress.Abort(progressId)
	} else {
		reader = reqContent
	}

	req, err := http.NewRequest("PUT", url, reader)
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}
	req.ContentLength = size
	req.Close = true

	setRequestHeaders(httpClientsDetails, size, req)
	setAuthentication(req, httpClientsDetails)
	addUserAgentHeader(req)

	client := jc.Client
	resp, err := client.Do(req)
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

// Read remote file,
// The caller is responsible to check if resp.StatusCode is StatusOK before reading, and to close io.ReadCloser after done reading.
func (jc *HttpClient) ReadRemoteFile(downloadPath string, httpClientsDetails httputils.HttpClientDetails) (io.ReadCloser, *http.Response, error) {
	resp, _, err := jc.sendGetForFileDownload(downloadPath, true, httpClientsDetails, 0)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp, nil
	}
	return resp.Body, resp, nil
}

// Bulk downloads a file.
// You may implement the log.Progress interface, or pass nil to run without progress display.
func (jc *HttpClient) DownloadFileWithProgress(downloadFileDetails *DownloadFileDetails, logMsgPrefix string,
	httpClientsDetails httputils.HttpClientDetails, retries int, isExplode bool, progress ioutils.Progress) (*http.Response, error) {
	resp, _, err := jc.downloadFile(downloadFileDetails, logMsgPrefix, true, httpClientsDetails, retries, isExplode, progress)
	return resp, err
}

// Bulk downloads a file.
func (jc *HttpClient) DownloadFile(downloadFileDetails *DownloadFileDetails, logMsgPrefix string,
	httpClientsDetails httputils.HttpClientDetails, retries int, isExplode bool) (*http.Response, error) {
	return jc.DownloadFileWithProgress(downloadFileDetails, logMsgPrefix, httpClientsDetails, retries, isExplode, nil)
}

func (jc *HttpClient) DownloadFileNoRedirect(downloadPath, localPath, fileName string, httpClientsDetails httputils.HttpClientDetails, retries int) (*http.Response, string, error) {
	downloadFileDetails := &DownloadFileDetails{DownloadPath: downloadPath, LocalPath: localPath, FileName: fileName}
	return jc.downloadFile(downloadFileDetails, "", false, httpClientsDetails, retries, false, nil)
}

func (jc *HttpClient) downloadFile(downloadFileDetails *DownloadFileDetails, logMsgPrefix string, followRedirect bool,
	httpClientsDetails httputils.HttpClientDetails, retries int, isExplode bool, progress ioutils.Progress) (resp *http.Response, redirectUrl string, err error) {
	retryExecutor := utils.RetryExecutor{
		MaxRetries:      retries,
		RetriesInterval: 0,
		ErrorMessage:    fmt.Sprintf("Failure occurred while downloading %s", downloadFileDetails.DownloadPath),
		LogMsgPrefix:    logMsgPrefix,
		ExecutionHandler: func() (bool, error) {
			resp, redirectUrl, err = jc.doDownloadFile(downloadFileDetails, logMsgPrefix, followRedirect, httpClientsDetails, isExplode, progress)
			// In case followRedirect is 'false' and doDownloadFile did redirect, an error is returned and redirectUrl
			// receives the redirect address. This case should not retry.
			if err != nil && !followRedirect && redirectUrl != "" {
				return false, err
			}
			// If error occurred during doDownloadFile, perform retry.
			if err != nil {
				return true, err
			}
			// Response must not be nil
			if resp == nil {
				return false, errorutils.CheckError(errors.New(fmt.Sprintf("%sReceived empty response from file download", logMsgPrefix)))
			}
			// If response-code < 500, should not retry
			if resp.StatusCode < 500 {
				return false, nil
			}
			// Perform retry
			log.Warn(fmt.Sprintf("%sArtifactory response: %s", logMsgPrefix, resp.Status))
			return true, nil
		},
	}

	err = retryExecutor.Execute()
	return
}

func (jc *HttpClient) doDownloadFile(downloadFileDetails *DownloadFileDetails, logMsgPrefix string, followRedirect bool,
	httpClientsDetails httputils.HttpClientDetails, isExplode bool, progress ioutils.Progress) (resp *http.Response, redirectUrl string, err error) {
	resp, redirectUrl, err = jc.sendGetForFileDownload(downloadFileDetails.DownloadPath, followRedirect, httpClientsDetails, 0)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp, redirectUrl, nil
	}

	isZip := fileutils.IsZip(downloadFileDetails.FileName)
	arch := archiver.MatchingFormat(downloadFileDetails.FileName)

	// If explode flag is true and the file is an archive but not zip, extract the file.
	if isExplode && !isZip && arch != nil {
		err = extractFile(downloadFileDetails, arch, resp.Body, logMsgPrefix)
		return
	}

	// Save the file to the file system
	err = saveToFile(downloadFileDetails, resp, progress)
	if err != nil {
		return
	}

	// Extract zip if necessary
	// Extracting zip after download to prevent out of memory issues.
	if isExplode && isZip {
		err = extractZip(downloadFileDetails, logMsgPrefix)
	}
	return
}

func saveToFile(downloadFileDetails *DownloadFileDetails, resp *http.Response, progress ioutils.Progress) error {
	fileName, err := fileutils.CreateFilePath(downloadFileDetails.LocalPath, downloadFileDetails.LocalFileName)
	if err != nil {
		return err
	}

	out, err := os.Create(fileName)
	if errorutils.CheckError(err) != nil {
		return err
	}

	defer out.Close()

	var reader io.Reader
	if progress != nil {
		progressId := progress.New(resp.ContentLength, "Downloading", downloadFileDetails.RelativePath)
		reader = progress.ReadWithProgress(progressId, resp.Body)
		defer progress.Abort(progressId)
	} else {
		reader = resp.Body
	}

	if len(downloadFileDetails.ExpectedSha1) > 0 {
		actualSha1 := sha1.New()
		writer := io.MultiWriter(actualSha1, out)

		_, err = io.Copy(writer, reader)
		if errorutils.CheckError(err) != nil {
			return err
		}

		if hex.EncodeToString(actualSha1.Sum(nil)) != downloadFileDetails.ExpectedSha1 {
			err = errors.New("Checksum mismatch for " + fileName + ", expected: " + downloadFileDetails.ExpectedSha1 + ", actual: " + hex.EncodeToString(actualSha1.Sum(nil)))
		}
	} else {
		_, err = io.Copy(out, reader)
	}

	return errorutils.CheckError(err)
}

func extractFile(downloadFileDetails *DownloadFileDetails, arch archiver.Archiver, reader io.Reader, logMsgPrefix string) error {
	log.Info(logMsgPrefix+"Extracting archive:", downloadFileDetails.FileName, "to", downloadFileDetails.LocalPath)
	err := fileutils.CreateDirIfNotExist(downloadFileDetails.LocalPath)
	if err != nil {
		return err
	}

	extractionPath, err := getExtractionPath(downloadFileDetails.LocalPath)
	if err != nil {
		return err
	}

	err = arch.Read(reader, extractionPath)
	return errorutils.CheckError(err)
}

func extractZip(downloadFileDetails *DownloadFileDetails, logMsgPrefix string) error {
	fileName, err := fileutils.CreateFilePath(downloadFileDetails.LocalPath, downloadFileDetails.LocalFileName)
	if err != nil {
		return err
	}
	log.Info(logMsgPrefix+"Extracting archive:", fileName, "to", downloadFileDetails.LocalPath)
	absLocalPath, err := filepath.Abs(downloadFileDetails.LocalPath)
	if errorutils.CheckError(err) != nil {
		return err
	}
	err = archiver.Zip.Open(fileName, absLocalPath)
	if errorutils.CheckError(err) != nil {
		return err
	}
	err = os.Remove(fileName)
	return errorutils.CheckError(err)
}

func getExtractionPath(localPath string) (string, error) {
	// The local path to which the file is going to be extracted,
	// needs to be absolute.
	absolutePath, err := filepath.Abs(localPath)
	if err != nil {
		return "", errorutils.CheckError(err)
	}
	// Add a trailing slash to the local path, since it has to be a directory.
	return absolutePath + string(os.PathSeparator), nil
}

// Downloads a file by chunks, concurrently.
// If successful, returns the resp of the last chunk, which will have resp.StatusCode = http.StatusPartialContent
// Otherwise: if an error occurred - returns the error with resp=nil, else - err=nil and the resp of the first chunk that received statusCode!=http.StatusPartialContent
// The caller is responsible to check the resp.StatusCode.
// You may implement the log.Progress interface, or pass nil to run without progress display.
func (jc *HttpClient) DownloadFileConcurrently(flags ConcurrentDownloadFlags, logMsgPrefix string,
	httpClientsDetails httputils.HttpClientDetails, progress ioutils.Progress) (*http.Response, error) {
	// Create temp dir for file chunks.
	tempDirPath, err := fileutils.CreateTempDir()
	if err != nil {
		return nil, err
	}
	defer fileutils.RemoveTempDir(tempDirPath)

	chunksPaths := make([]string, flags.SplitCount)

	var downloadProgressId int
	if progress != nil {
		downloadProgressId = progress.New(flags.FileSize, "Downloading", flags.RelativePath)
		mergingProgressId := progress.NewReplacement(downloadProgressId, "  Merging  ", flags.RelativePath)
		// Aborting order matters. mergingProgress depends on the existence of downloadingProgress
		defer progress.Abort(downloadProgressId)
		defer progress.Abort(mergingProgressId)
	}

	resp, err := jc.downloadChunksConcurrently(chunksPaths, flags, logMsgPrefix, tempDirPath, httpClientsDetails, progress, downloadProgressId)
	if err != nil {
		return nil, err
	}
	// If not all chunks were downloaded successfully, return
	if resp.StatusCode != http.StatusPartialContent {
		return resp, nil
	}

	if flags.LocalPath != "" {
		err = os.MkdirAll(flags.LocalPath, 0777)
		if errorutils.CheckError(err) != nil {
			return nil, err
		}
		flags.LocalFileName = filepath.Join(flags.LocalPath, flags.LocalFileName)
	}

	if fileutils.IsPathExists(flags.LocalFileName, false) {
		err := os.Remove(flags.LocalFileName)
		if errorutils.CheckError(err) != nil {
			return nil, err
		}
	}

	// Explode and merge archive if necessary
	if flags.Explode {
		extracted, err := extractAndMergeChunks(chunksPaths, flags, logMsgPrefix)
		if err != nil {
			return nil, err
		}
		if extracted {
			return resp, nil
		}
	}

	err = mergeChunks(chunksPaths, flags)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	log.Info(logMsgPrefix + "Done downloading.")
	return resp, nil
}

// The caller is responsible to check that resp.StatusCode is http.StatusOK
func (jc *HttpClient) GetRemoteFileDetails(downloadUrl string, httpClientsDetails httputils.HttpClientDetails) (*fileutils.FileDetails, *http.Response, error) {
	resp, _, err := jc.SendHead(downloadUrl, httpClientsDetails)
	if err != nil {
		return nil, nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, resp, nil
	}

	fileSize := int64(0)
	contentLength := resp.Header.Get("Content-Length")
	if len(contentLength) > 0 {
		fileSize, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return nil, nil, err
		}
	}

	fileDetails := new(fileutils.FileDetails)
	fileDetails.Checksum.Md5 = resp.Header.Get("X-Checksum-Md5")
	fileDetails.Checksum.Sha1 = resp.Header.Get("X-Checksum-Sha1")
	fileDetails.Size = fileSize
	return fileDetails, resp, nil
}

// Downloads chunks, concurrently.
// If successful, returns the resp of the last chunk, which will have resp.StatusCode = http.StatusPartialContent
// Otherwise: if an error occurred - returns the error with resp=nil, else - err=nil and the resp of the first chunk that received statusCode!=http.StatusPartialContent
// The caller is responsible to check the resp.StatusCode.
func (jc *HttpClient) downloadChunksConcurrently(chunksPaths []string, flags ConcurrentDownloadFlags, logMsgPrefix,
	chunksDownloadPath string, httpClientsDetails httputils.HttpClientDetails, progress ioutils.Progress, progressId int) (*http.Response, error) {
	var wg sync.WaitGroup
	chunkSize := flags.FileSize / int64(flags.SplitCount)
	mod := flags.FileSize % int64(flags.SplitCount)
	// Create a list of errors, to allow each go routine to save there its own returned error.
	errorsList := make([]error, flags.SplitCount)
	// Store the responses, to return a response with unexpected statusCode or the last response if all successful
	respList := make([]*http.Response, flags.SplitCount)
	// Global vars on top of the go routines, to break the loop earlier if needed
	var err error
	var resp *http.Response
	for i := 0; i < flags.SplitCount; i++ {
		// Checking this global error may help break out of the loop earlier, if an error or the wrong status code was received
		// has already been returned by one of the go routines.
		if err != nil {
			break
		}
		if resp != nil && resp.StatusCode != http.StatusPartialContent {
			break
		}
		wg.Add(1)
		start := chunkSize * int64(i)
		end := chunkSize * (int64(i) + 1)
		if i == flags.SplitCount-1 {
			end += mod
		}
		requestClientDetails := httpClientsDetails.Clone()
		go func(start, end int64, i int) {
			chunksPaths[i], respList[i], errorsList[i] = jc.downloadFileRange(flags, start, end, i, logMsgPrefix, chunksDownloadPath, *requestClientDetails, flags.Retries, progress, progressId)
			// Write to the global vars if the chunk wasn't downloaded successfully
			if errorsList[i] != nil {
				err = errorsList[i]
			}
			if respList[i] != nil && respList[i].StatusCode != http.StatusPartialContent {
				resp = respList[i]
			}
			wg.Done()
		}(start, end, i)
	}
	wg.Wait()

	// Verify that all chunks have been downloaded successfully.
	for _, e := range errorsList {
		if e != nil {
			return nil, errorutils.CheckError(e)
		}
	}
	for _, r := range respList {
		if r.StatusCode != http.StatusPartialContent {
			return r, nil
		}
	}

	// If all chunks were downloaded successfully, return the response of the last chunk.
	return respList[len(respList)-1], nil
}

func mergeChunks(chunksPaths []string, flags ConcurrentDownloadFlags) error {
	destFile, err := os.OpenFile(flags.LocalFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if errorutils.CheckError(err) != nil {
		return err
	}
	defer destFile.Close()
	var writer io.Writer
	var actualSha1 hash.Hash
	if len(flags.ExpectedSha1) > 0 {
		actualSha1 = sha1.New()
		writer = io.MultiWriter(actualSha1, destFile)
	} else {
		writer = io.MultiWriter(destFile)
	}
	for i := 0; i < flags.SplitCount; i++ {
		reader, err := os.Open(chunksPaths[i])
		if err != nil {
			return err
		}
		defer reader.Close()
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
	}
	if len(flags.ExpectedSha1) > 0 {
		if hex.EncodeToString(actualSha1.Sum(nil)) != flags.ExpectedSha1 {
			err = errors.New("Checksum mismatch for  " + flags.LocalFileName + ", expected: " + flags.ExpectedSha1 + ", actual: " + hex.EncodeToString(actualSha1.Sum(nil)))
		}
	}
	return err
}

func extractAndMergeChunks(chunksPaths []string, flags ConcurrentDownloadFlags, logMsgPrefix string) (bool, error) {
	if fileutils.IsZip(flags.FileName) {
		multiReader, err := ioutils.NewMultiFileReaderAt(chunksPaths)
		if errorutils.CheckError(err) != nil {
			return false, err
		}
		log.Info(logMsgPrefix+"Extracting archive:", flags.FileName, "to", flags.LocalPath)
		err = fileutils.Unzip(multiReader, multiReader.Size(), flags.LocalPath)
		if errorutils.CheckError(err) != nil {
			return false, err
		}
		return true, nil
	}

	arch := archiver.MatchingFormat(flags.FileName)
	if arch == nil {
		log.Debug(logMsgPrefix+"Not an archive:", flags.FileName, "downloading file without extracting it.")
		return false, nil
	}

	fileReaders := make([]io.Reader, len(chunksPaths))
	var err error
	for k, v := range chunksPaths {
		f, err := os.Open(v)
		fileReaders[k] = f
		if err != nil {
			return false, errorutils.CheckError(err)
		}
		defer f.Close()
	}

	multiReader := io.MultiReader(fileReaders...)
	extractionPath, err := getExtractionPath(flags.LocalPath)
	if err != nil {
		return false, err
	}
	log.Info(logMsgPrefix+"Extracting archive:", flags.FileName, "to", extractionPath)
	err = arch.Read(multiReader, extractionPath)
	if err != nil {
		return false, errorutils.CheckError(err)
	}
	return true, nil
}

func (jc *HttpClient) downloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix, chunkDownloadPath string,
	httpClientsDetails httputils.HttpClientDetails, retries int, progress ioutils.Progress, progressId int) (fileName string, resp *http.Response, err error) {
	retryExecutor := utils.RetryExecutor{
		MaxRetries:      retries,
		RetriesInterval: 0,
		ErrorMessage:    fmt.Sprintf("Failure occurred while downloading part %d of %s", currentSplit, flags.DownloadPath),
		LogMsgPrefix:    fmt.Sprintf("%s[%s]: ", logMsgPrefix, strconv.Itoa(currentSplit)),
		ExecutionHandler: func() (bool, error) {
			fileName, resp, err = jc.doDownloadFileRange(flags, start, end, currentSplit, logMsgPrefix, chunkDownloadPath, httpClientsDetails, progress, progressId)
			if err != nil {
				return true, err
			}
			// Response must not be nil
			if resp == nil {
				return false, errorutils.CheckError(errors.New(fmt.Sprintf("%s[%s]: Received empty response from file download", logMsgPrefix, strconv.Itoa(currentSplit))))
			}
			// If response-code < 500, should not retry
			if resp.StatusCode < 500 {
				return false, nil
			}
			// Perform retry
			log.Warn(fmt.Sprintf("%s[%s]: Artifactory response: %s", logMsgPrefix, strconv.Itoa(currentSplit), resp.Status))
			return true, nil
		},
	}

	err = retryExecutor.Execute()
	return
}

func (jc *HttpClient) doDownloadFileRange(flags ConcurrentDownloadFlags, start, end int64, currentSplit int, logMsgPrefix, chunkDownloadPath string,
	httpClientsDetails httputils.HttpClientDetails, progress ioutils.Progress, progressId int) (fileName string, resp *http.Response, err error) {

	tempFile, err := ioutil.TempFile(chunkDownloadPath, strconv.Itoa(currentSplit)+"_")
	if errorutils.CheckError(err) != nil {
		return
	}
	defer tempFile.Close()

	if httpClientsDetails.Headers == nil {
		httpClientsDetails.Headers = make(map[string]string)
	}
	httpClientsDetails.Headers["Range"] = "bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)
	resp, _, err = jc.sendGetForFileDownload(flags.DownloadPath, true, httpClientsDetails, currentSplit)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	// Unexpected http response
	if resp.StatusCode != http.StatusPartialContent {
		return
	}
	log.Info(fmt.Sprintf("%s[%s]: %s...", logMsgPrefix, strconv.Itoa(currentSplit), resp.Status))

	err = os.MkdirAll(chunkDownloadPath, 0777)
	if errorutils.CheckError(err) != nil {
		return "", nil, err
	}

	var reader io.Reader
	if progress != nil {
		reader = progress.ReadWithProgress(progressId, resp.Body)
	} else {
		reader = resp.Body
	}

	_, err = io.Copy(tempFile, reader)

	if errorutils.CheckError(err) != nil {
		return "", nil, err
	}
	return tempFile.Name(), resp, errorutils.CheckError(err)
}

// The caller is responsible to check if resp.StatusCode is StatusOK before relying on the bool value
func (jc *HttpClient) IsAcceptRanges(downloadUrl string, httpClientsDetails httputils.HttpClientDetails) (bool, *http.Response, error) {
	resp, _, err := jc.SendHead(downloadUrl, httpClientsDetails)
	if errorutils.CheckError(err) != nil {
		return false, nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, resp, nil
	}
	return resp.Header.Get("Accept-Ranges") == "bytes", resp, nil
}

func setAuthentication(req *http.Request, httpClientsDetails httputils.HttpClientDetails) {
	//Set authentication
	if httpClientsDetails.ApiKey != "" {
		if httpClientsDetails.User != "" {
			req.SetBasicAuth(httpClientsDetails.User, httpClientsDetails.ApiKey)
		} else {
			req.Header.Set("X-JFrog-Art-Api", httpClientsDetails.ApiKey)
		}
		return
	}
	if httpClientsDetails.AccessToken != "" {
		if httpClientsDetails.User != "" {
			req.SetBasicAuth(httpClientsDetails.User, httpClientsDetails.AccessToken)
		} else {
			req.Header.Set("Authorization", "Bearer "+httpClientsDetails.AccessToken)
		}
		return
	}
	if httpClientsDetails.Password != "" {
		req.SetBasicAuth(httpClientsDetails.User, httpClientsDetails.Password)
	}
}

func addUserAgentHeader(req *http.Request) {
	req.Header.Set("User-Agent", utils.GetUserAgent())
}

type DownloadFileDetails struct {
	FileName      string `json:"LocalFileName,omitempty"`
	DownloadPath  string `json:"DownloadPath,omitempty"`
	RelativePath  string `json:"RelativePath,omitempty"`
	LocalPath     string `json:"LocalPath,omitempty"`
	LocalFileName string `json:"LocalFileName,omitempty"`
	ExpectedSha1  string `json:"ExpectedSha1,omitempty"`
	Size          int64  `json:"Size,omitempty"`
}

type ConcurrentDownloadFlags struct {
	FileName      string
	DownloadPath  string
	RelativePath  string
	LocalFileName string
	LocalPath     string
	ExpectedSha1  string
	FileSize      int64
	SplitCount    int
	Explode       bool
	Retries       int
}
