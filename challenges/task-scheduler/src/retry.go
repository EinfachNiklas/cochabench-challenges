package taskscheduler

import (
	"time"
)

func CalculateDelay(attempt int, policy RetryPolicy) time.Duration {
	return 0
}

func IsRetryable(err error) bool {
	return false
}

func WrapRetryable(err error) error {
	return err
}

func WrapNonRetryable(err error) error {
	return err
}

func NewRetryHandler(policy RetryPolicy, handler TaskHandler) TaskHandler {
	return handler
}
