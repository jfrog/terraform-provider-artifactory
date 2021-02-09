package errorutils

import (
	"errors"
	"io/ioutil"
	"net/http"
)

// Error modes (how should the application behave when the CheckError function is invoked):
type OnErrorHandler func(error) error

var CheckError = func(err error) error {
	return err
}

// Check expected status codes and return error if needed
func CheckResponseStatus(resp *http.Response, expectedStatusCodes ...int) error {
	for _, statusCode := range expectedStatusCodes {
		if statusCode == resp.StatusCode {
			return nil
		}
	}

	errorBody, _ := ioutil.ReadAll(resp.Body)
	return errors.New(resp.Status + " " + string(errorBody))
}
