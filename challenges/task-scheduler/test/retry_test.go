package taskscheduler

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestCalculateDelay_BasicBackoff(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  time.Second,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
	}

	expected := []time.Duration{
		1 * time.Second,  // attempt 0: 1 * 2^0 = 1s
		2 * time.Second,  // attempt 1: 1 * 2^1 = 2s
		4 * time.Second,  // attempt 2: 1 * 2^2 = 4s
		8 * time.Second,  // attempt 3: 1 * 2^3 = 8s
		16 * time.Second, // attempt 4: 1 * 2^4 = 16s
	}

	for i, exp := range expected {
		got := CalculateDelay(i, policy)
		// Erlaube kleine Rundungsfehler
		diff := got - exp
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Millisecond {
			t.Errorf("Attempt %d: expected %v, got %v", i, exp, got)
		}
	}
}

func TestCalculateDelay_MaxDelayCap(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 10,
		BaseDelay:  time.Second,
		MaxDelay:   5 * time.Second,
		Multiplier: 3.0,
	}

	// Attempt 3: 1 * 3^3 = 27s → gekappt auf 5s
	delay := CalculateDelay(3, policy)
	if delay > 5*time.Second+time.Millisecond {
		t.Errorf("Expected delay capped at 5s, got %v", delay)
	}
}

func TestCalculateDelay_Multiplier(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   10 * time.Second,
		Multiplier: 1.5,
	}

	// Attempt 0: 100ms * 1.5^0 = 100ms
	delay0 := CalculateDelay(0, policy)
	if delay0 != 100*time.Millisecond {
		t.Errorf("Attempt 0: expected 100ms, got %v", delay0)
	}

	// Attempt 1: 100ms * 1.5^1 = 150ms
	delay1 := CalculateDelay(1, policy)
	expected := time.Duration(float64(100*time.Millisecond) * 1.5)
	diff := delay1 - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Millisecond {
		t.Errorf("Attempt 1: expected ~%v, got %v", expected, delay1)
	}
}

func TestIsRetryable_RetryableError(t *testing.T) {
	err := &RetryableError{Err: errors.New("connection timeout")}
	if !IsRetryable(err) {
		t.Error("RetryableError should be retryable")
	}
}

func TestIsRetryable_NonRetryableError(t *testing.T) {
	err := &NonRetryableError{Err: errors.New("invalid input")}
	if IsRetryable(err) {
		t.Error("NonRetryableError should NOT be retryable")
	}
}

func TestIsRetryable_GenericError(t *testing.T) {
	err := errors.New("unknown error")
	if !IsRetryable(err) {
		t.Error("Generic errors should be retryable by default")
	}
}

func TestIsRetryable_NilError(t *testing.T) {
	if IsRetryable(nil) {
		t.Error("nil error should not be retryable")
	}
}

func TestIsRetryable_WrappedRetryableError(t *testing.T) {
	inner := &RetryableError{Err: errors.New("timeout")}
	wrapped := fmt.Errorf("request failed: %w", inner)
	if !IsRetryable(wrapped) {
		t.Error("Wrapped RetryableError should be retryable")
	}
}

func TestIsRetryable_WrappedNonRetryableError(t *testing.T) {
	inner := &NonRetryableError{Err: errors.New("bad request")}
	wrapped := fmt.Errorf("request failed: %w", inner)
	if IsRetryable(wrapped) {
		t.Error("Wrapped NonRetryableError should NOT be retryable")
	}
}

func TestWrapRetryable(t *testing.T) {
	original := errors.New("timeout")
	wrapped := WrapRetryable(original)

	var retryErr *RetryableError
	if !errors.As(wrapped, &retryErr) {
		t.Error("WrapRetryable should return a *RetryableError")
	}
	if retryErr.Err != original {
		t.Error("WrapRetryable should preserve original error")
	}
}

func TestWrapNonRetryable(t *testing.T) {
	original := errors.New("invalid input")
	wrapped := WrapNonRetryable(original)

	var nonRetryErr *NonRetryableError
	if !errors.As(wrapped, &nonRetryErr) {
		t.Error("WrapNonRetryable should return a *NonRetryableError")
	}
	if nonRetryErr.Err != original {
		t.Error("WrapNonRetryable should preserve original error")
	}
}

func TestNewRetryHandler_SuccessOnFirstTry(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
		Multiplier: 2.0,
	}

	callCount := 0
	handler := func(task *Task) error {
		callCount++
		return nil
	}

	retryHandler := NewRetryHandler(policy, handler)
	task := &Task{ID: "task-1", MaxRetries: 3}

	err := retryHandler(task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestNewRetryHandler_RetryOnFailure(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Multiplier: 2.0,
	}

	callCount := 0
	handler := func(task *Task) error {
		callCount++
		if callCount < 3 {
			return &RetryableError{Err: errors.New("temporary failure")}
		}
		return nil // Dritter Versuch erfolgreich
	}

	retryHandler := NewRetryHandler(policy, handler)
	task := &Task{ID: "task-1", MaxRetries: 3}

	err := retryHandler(task)
	if err != nil {
		t.Fatalf("Expected success after retries, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("Expected 3 calls (2 retries + 1 success), got %d", callCount)
	}
	if task.RetryCount != 2 {
		t.Errorf("Expected RetryCount = 2, got %d", task.RetryCount)
	}
}

func TestNewRetryHandler_MaxRetriesExceeded(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 2,
		BaseDelay:  time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Multiplier: 2.0,
	}

	callCount := 0
	handler := func(task *Task) error {
		callCount++
		return &RetryableError{Err: errors.New("always fails")}
	}

	retryHandler := NewRetryHandler(policy, handler)
	task := &Task{ID: "task-1", MaxRetries: 2}

	err := retryHandler(task)
	if err == nil {
		t.Fatal("Expected error after max retries")
	}
	// 1 initial + 2 retries = 3 calls
	if callCount != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount)
	}
}

func TestNewRetryHandler_NonRetryableError_NoRetry(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 5,
		BaseDelay:  time.Millisecond,
		MaxDelay:   time.Second,
		Multiplier: 2.0,
	}

	callCount := 0
	handler := func(task *Task) error {
		callCount++
		return &NonRetryableError{Err: errors.New("permanent failure")}
	}

	retryHandler := NewRetryHandler(policy, handler)
	task := &Task{ID: "task-1", MaxRetries: 5}

	err := retryHandler(task)
	if err == nil {
		t.Fatal("Expected error")
	}
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retry for NonRetryableError), got %d", callCount)
	}
}

func TestNewRetryHandler_SetsRetryStatus(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries: 3,
		BaseDelay:  time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
		Multiplier: 2.0,
	}

	statuses := []TaskStatus{}
	handler := func(task *Task) error {
		statuses = append(statuses, task.Status)
		if len(statuses) < 2 {
			return &RetryableError{Err: errors.New("fail")}
		}
		return nil
	}

	retryHandler := NewRetryHandler(policy, handler)
	task := &Task{ID: "task-1", MaxRetries: 3, Status: StatusRunning}

	retryHandler(task)

	// The second call should occur with StatusRetrying
	if len(statuses) < 2 {
		t.Fatal("Expected at least 2 handler calls")
	}
	if statuses[1] != StatusRetrying {
		t.Errorf("Expected StatusRetrying for retry attempt, got %v", statuses[1])
	}
}
