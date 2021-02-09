package content

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// Open and read JSON file, find the array key inside it and load its value into the memory in small chunks.
// Currently, 'ContentReader' only support extracting a single value for a given key (arrayKey), other keys are ignored.
// The value must of of type array.
// Each array value can be fetched using 'GetRecord' (thread-safe).
// This technique solves the limit of memory size which may be too small to fit large JSON.
type ContentReader struct {
	// filePath - source data file path.
	// arrayKey = Read the value of the specific object in JSON.
	filePath, arrayKey string
	// The objects from the source data file are being pushed into the data channel.
	dataChannel chan map[string]interface{}
	errorsQueue *utils.ErrorsQueue
	once        *sync.Once
	// Number of element in the array (cache)
	length int
	empty  bool
}

func NewContentReader(filePath string, arrayKey string) *ContentReader {
	self := ContentReader{}
	self.filePath = filePath
	self.arrayKey = arrayKey
	self.dataChannel = make(chan map[string]interface{}, utils.MaxBufferSize)
	self.errorsQueue = utils.NewErrorsQueue(utils.MaxBufferSize)
	self.once = new(sync.Once)
	self.empty = filePath == ""
	return &self
}

func NewEmptyContentReader(arrayKey string) *ContentReader {
	self := NewContentReader("", arrayKey)
	return self
}

func (cr *ContentReader) IsEmpty() bool {
	return cr.empty
}

// Each call to 'NextRecord()' will returns a single element from the channel.
// Only the first call invokes a goroutine to read data from the file and push it into the channel.
// 'io.EOF' will be returned if no data is left.
func (cr *ContentReader) NextRecord(recordOutput interface{}) error {
	if cr.empty {
		return errorutils.CheckError(errors.New("Empty"))
	}
	cr.once.Do(func() {
		go func() {
			defer close(cr.dataChannel)
			cr.length = 0
			cr.run()
		}()
	})
	record, ok := <-cr.dataChannel
	if !ok {
		return errorutils.CheckError(io.EOF)
	}
	// Transform the data into a Go type
	data, err := json.Marshal(record)
	if err != nil {
		cr.errorsQueue.AddError(errorutils.CheckError(err))
		return err
	}
	err = errorutils.CheckError(json.Unmarshal(data, recordOutput))
	if err != nil {
		cr.errorsQueue.AddError(err)
	}
	cr.length++
	return err
}

// Prepare the reader to read the file all over again (not thread-safe).
func (cr *ContentReader) Reset() {
	cr.dataChannel = make(chan map[string]interface{}, utils.MaxBufferSize)
	cr.once = new(sync.Once)
}

// Cleanup the reader data.
func (cr *ContentReader) Close() error {
	if cr.filePath != "" {
		if err := errorutils.CheckError(os.Remove(cr.filePath)); err != nil {
			log.Error(err)
			return err
		}
		cr.filePath = ""
	}
	return nil
}

func (cr *ContentReader) GetFilePath() string {
	return cr.filePath
}

// Number of element in the array.
func (cr *ContentReader) Length() (int, error) {
	if cr.empty == true {
		return 0, nil
	}
	if cr.length == 0 {
		for item := new(interface{}); cr.NextRecord(item) == nil; item = new(interface{}) {
		}
		cr.Reset()
		if err := cr.GetError(); err != nil {
			return 0, err
		}
	}
	return cr.length, nil
}

// Open and read the file. Push each array element into the channel.
// The channel may block the thread, therefore should run async.
func (cr *ContentReader) run() {
	fd, err := os.Open(cr.filePath)
	if err != nil {
		log.Error(err.Error())
		cr.errorsQueue.AddError(errorutils.CheckError(err))
		return
	}
	defer fd.Close()
	br := bufio.NewReaderSize(fd, 65536)
	dec := json.NewDecoder(br)
	err = findDecoderTargetPosition(dec, cr.arrayKey, true)
	if err != nil {
		if err == io.EOF {
			cr.errorsQueue.AddError(errorutils.CheckError(errors.New(cr.arrayKey + " not found")))
			return
		}
		cr.errorsQueue.AddError(err)
		log.Error(err.Error())
		return
	}
	for dec.More() {
		var ResultItem map[string]interface{}
		err := dec.Decode(&ResultItem)
		if err != nil {
			log.Error(err)
			cr.errorsQueue.AddError(errorutils.CheckError(err))
			return
		}
		cr.dataChannel <- ResultItem
	}
}

func (cr *ContentReader) GetError() error {
	return cr.errorsQueue.GetError()
}

// Search and set the decoder's position at the desired key in the JSON file.
// If the desired key is not found, return io.EOF
func findDecoderTargetPosition(dec *json.Decoder, target string, isArray bool) error {
	for dec.More() {
		// Token returns the next JSON token in the input stream.
		t, err := dec.Token()
		if err != nil {
			return errorutils.CheckError(err)
		}
		if t == target {
			if isArray {
				// Skip '['
				_, err = dec.Token()
			}
			return errorutils.CheckError(err)
		}
	}
	return nil
}

// Scan the JSON file and check if the array contains at least one element.
func isEmptyArray(dec *json.Decoder, target string, isArray bool) (bool, error) {
	if err := findDecoderTargetPosition(dec, target, isArray); err != nil {
		return false, err
	}
	t, err := dec.Token()
	if err != nil {
		return false, errorutils.CheckError(err)
	}
	return t == json.Delim('{'), nil
}

func MergeReaders(arr []*ContentReader, arrayKey string) (*ContentReader, error) {
	cw, err := NewContentWriter(arrayKey, true, false)
	if err != nil {
		return nil, err
	}
	defer cw.Close()
	for _, cr := range arr {
		for item := new(interface{}); cr.NextRecord(item) == nil; item = new(interface{}) {
			cw.Write(*item)
		}
		if err := cr.GetError(); err != nil {
			return nil, err
		}
	}
	return NewContentReader(cw.GetFilePath(), arrayKey), nil
}
