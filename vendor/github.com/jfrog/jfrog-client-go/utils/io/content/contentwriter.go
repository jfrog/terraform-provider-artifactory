package content

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

const (
	jsonArrayPrefixPattern = "  \"%s\": ["
	jsonArraySuffix        = "]\n"
	DefaultKey             = "results"
)

// Write a JSON file in small chunks. Only a single JSON key can be written to the file, and array as its value.
// The array's values could be any JSON value types (number, string, etc...).
// Once the first 'Write" call is made, the file will stay open, waiting for the next struct to be written (thread-safe).
// Finally, 'Close' will fill the end of the JSON file and the operation will be completed.
type ContentWriter struct {
	// arrayKey = JSON object key to be written.
	arrayKey string
	// The output data file path.
	outputFile *os.File
	// The chanel which from the output records will be pulled.
	dataChannel    chan interface{}
	isCompleteFile bool
	errorsQueue    *utils.ErrorsQueue
	runWaiter      sync.WaitGroup
	once           sync.Once
	empty          bool
	useStdout      bool
}

func NewContentWriter(arrayKey string, isCompleteFile, useStdout bool) (*ContentWriter, error) {
	self := ContentWriter{}
	self.useStdout = useStdout
	self.arrayKey = arrayKey
	self.dataChannel = make(chan interface{}, utils.MaxBufferSize)
	self.errorsQueue = utils.NewErrorsQueue(utils.MaxBufferSize)
	self.isCompleteFile = isCompleteFile
	self.empty = true
	return &self, nil
}

func (rw *ContentWriter) SetArrayKey(arrKey string) *ContentWriter {
	rw.arrayKey = arrKey
	return rw
}

func (rw *ContentWriter) GetArrayKey() string {
	return rw.arrayKey
}

func (rw *ContentWriter) IsEmpty() bool {
	return rw.empty
}

func (rw *ContentWriter) GetFilePath() string {
	if rw.outputFile != nil {
		return rw.outputFile.Name()
	}
	return ""
}

func (rw *ContentWriter) RemoveOutputFilePath() error {
	return errorutils.CheckError(os.Remove(rw.outputFile.Name()))
}

// Write a single item to the JSON array.
func (rw *ContentWriter) Write(record interface{}) {
	rw.empty = false
	rw.startWritingWorker()
	rw.dataChannel <- record
}

func (rw *ContentWriter) startWritingWorker() {
	rw.once.Do(func() {
		var err error
		if rw.useStdout {
			rw.outputFile = os.Stdout
		} else {
			rw.outputFile, err = fileutils.CreateTempFile()
			if err != nil {
				rw.errorsQueue.AddError(errorutils.CheckError(err))
				return
			}
		}
		rw.runWaiter.Add(1)
		go func() {
			defer rw.runWaiter.Done()
			rw.run()
		}()
	})
}

// Write the data from the channel to JSON file.
// The channel may block the thread, therefore should run async.
func (rw *ContentWriter) run() {
	var err error
	if !rw.useStdout {
		defer rw.outputFile.Close()
	}
	openString := jsonArrayPrefixPattern
	closeString := ""
	if rw.isCompleteFile {
		openString = "{\n" + openString
	}
	_, err = rw.outputFile.WriteString(fmt.Sprintf(openString, rw.arrayKey))
	if err != nil {
		rw.errorsQueue.AddError(errorutils.CheckError(err))
		return
	}
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("    ", "  ")
	recordPrefix := "\n    "
	firstRecord := true
	for record := range rw.dataChannel {
		buf.Reset()
		err = enc.Encode(record)
		if err != nil {
			rw.errorsQueue.AddError(errorutils.CheckError(err))
			continue
		}
		record := recordPrefix + string(bytes.TrimRight(buf.Bytes(), "\n"))
		_, err = rw.outputFile.WriteString(record)
		if err != nil {
			rw.errorsQueue.AddError(errorutils.CheckError(err))
			continue
		}
		if firstRecord {
			// If a record was printed, we want to print a comma and ne before each and every future record.
			recordPrefix = "," + recordPrefix
			// We will close the array in a new-indent line.
			closeString = "\n  "
			firstRecord = false
		}
	}
	closeString = closeString + jsonArraySuffix
	if rw.isCompleteFile {
		closeString += "}\n"
	}
	_, err = rw.outputFile.WriteString(closeString)
	if err != nil {
		rw.errorsQueue.AddError(errorutils.CheckError(err))
	}
	return
}

// Finish writing to the file.
func (rw *ContentWriter) Close() error {
	if rw.empty {
		return nil
	}
	close(rw.dataChannel)
	rw.runWaiter.Wait()
	if err := rw.GetError(); err != nil {
		log.Error("Failed to close writer: " + err.Error())
		return err
	}
	return nil
}

func (rw *ContentWriter) GetError() error {
	return rw.errorsQueue.GetError()
}
