package utils

import (
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"time"
)

type ExecutionHandlerFunc func() (bool, error)

type RetryExecutor struct {
	// The amount of retries to perform.
	MaxRetries int

	// Number of seconds to sleep between retries.
	RetriesInterval int

	// Message to display when retrying.
	ErrorMessage string

	// Prefix to print at the beginning of each log.
	LogMsgPrefix string

	// ExecutionHandler is the operation to run with retries.
	ExecutionHandler ExecutionHandlerFunc
}

func (runner *RetryExecutor) Execute() error {
	var err error
	var shouldRetry bool
	for i := 0; i <= runner.MaxRetries; i++ {
		// Run ExecutionHandler
		shouldRetry, err = runner.ExecutionHandler()

		// If should not retry, return
		if !shouldRetry {
			return err
		}

		log.Warn(runner.getLogRetryMessage(i, err))
		// Going to sleep for RetryInterval seconds
		if runner.RetriesInterval > 0 && i < runner.MaxRetries {
			time.Sleep(time.Second * time.Duration(runner.RetriesInterval))
		}
	}

	return err
}

func (runner *RetryExecutor) getLogRetryMessage(attemptNumber int, err error) (message string) {
	message = fmt.Sprintf("%sAttempt %v - %s", runner.LogMsgPrefix, attemptNumber, runner.ErrorMessage)
	if err != nil {
		message = fmt.Sprintf("%s - %s", message, err.Error())
	}
	return
}
