package client

import (
	"fmt"
	"net/http"
)

// ErrorResponse reports one or more errors caused by an API request.
type ErrorResponse struct {
	Response *http.Response `json:"-"`                // HTTP response that caused this error
	Errors   []Status       `json:"errors,omitempty"` // Individual errors
}

// Status is the individual error provided by the API
type Status struct {
	Status  int    `json:"status"`  // Validation error status code
	Message string `json:"message"` // Message describing the error. Errors with Code == "custom" will always have this set.
}

func (e *Status) Error() string {
	return fmt.Sprintf("%d error caused by %s", e.Status, e.Message)
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %+v", r.Response.Request.Method, r.Response.Request.URL,
		r.Response.StatusCode, r.Errors)
}
