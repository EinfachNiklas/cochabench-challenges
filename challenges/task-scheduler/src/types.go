package taskscheduler

import (
	"time"
)

// Priority defines the priority levels for tasks
type Priority int

const (
	PriorityLow      Priority = 0
	PriorityMedium   Priority = 1
	PriorityHigh     Priority = 2
	PriorityCritical Priority = 3
)

// String returns the priority as a human-readable string
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "LOW"
	case PriorityMedium:
		return "MEDIUM"
	case PriorityHigh:
		return "HIGH"
	case PriorityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// IsValid checks whether the priority is valid
func (p Priority) IsValid() bool {
	return p >= PriorityLow && p <= PriorityCritical
}

// TaskStatus represents the lifecycle status of a task
type TaskStatus int

const (
	StatusPending   TaskStatus = 0
	StatusRunning   TaskStatus = 1
	StatusCompleted TaskStatus = 2
	StatusFailed    TaskStatus = 3
	StatusRetrying  TaskStatus = 4
	StatusCancelled TaskStatus = 5
)

// String returns the status as a human-readable string
func (s TaskStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusRunning:
		return "RUNNING"
	case StatusCompleted:
		return "COMPLETED"
	case StatusFailed:
		return "FAILED"
	case StatusRetrying:
		return "RETRYING"
	case StatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// IsTerminal returns true if the status is a terminal state
func (s TaskStatus) IsTerminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// Task represents a schedulable unit of work
type Task struct {
	ID           string                 // Unique task ID
	Name         string                 // Human-readable name
	Priority     Priority               // Priority for execution order
	Status       TaskStatus             // Current lifecycle status
	Handler      string                 // Name of the registered handler
	Payload      map[string]interface{} // Arbitrary data for the handler
	Dependencies []string               // Task IDs that must complete first
	MaxRetries   int                    // Maximum number of retry attempts
	RetryCount   int                    // Current number of retry attempts
	CreatedAt    time.Time              // Creation timestamp
	StartedAt    time.Time              // Execution start timestamp
	CompletedAt  time.Time              // Completion timestamp
	Error        error                  // Last error (if any)
	Result       interface{}            // Result of the execution
}

// TaskHandler is a function that processes a task
type TaskHandler func(task *Task) error

// RetryPolicy defines the retry behavior for failed tasks
type RetryPolicy struct {
	MaxRetries int           // Maximum number of retry attempts
	BaseDelay  time.Duration // Base wait time between attempts
	MaxDelay   time.Duration // Maximum wait time (cap for backoff)
	Multiplier float64       // Multiplier for exponential backoff
}

// SchedulerConfig holds the configuration for the scheduler
type SchedulerConfig struct {
	MaxWorkers      int           // Maximum number of concurrent workers
	QueueSize       int           // Capacity of the task queue
	DefaultRetry    RetryPolicy   // Default retry policy
	ShutdownTimeout time.Duration // Timeout for graceful shutdown
}

// TaskFilter allows filtering tasks in the store
type TaskFilter struct {
	Status   *TaskStatus // Filter by status (nil = all)
	Priority *Priority   // Filter by priority (nil = all)
	Handler  string      // Filter by handler name (empty = all)
}

// TaskMetrics holds statistics about task executions
type TaskMetrics struct {
	TotalSubmitted  int64                // Total number of submitted tasks
	TotalCompleted  int64                // Total number of completed tasks
	TotalFailed     int64                // Total number of failed tasks
	TotalRetried    int64                // Total number of retried tasks
	AverageExecTime time.Duration        // Average execution time
	TasksByPriority map[Priority]int64   // Number of tasks per priority
	TasksByStatus   map[TaskStatus]int64 // Number of tasks per status
}

// RetryableError signals an error that should trigger a retry
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NonRetryableError signals an error that should NOT trigger a retry
type NonRetryableError struct {
	Err error
}

func (e *NonRetryableError) Error() string {
	return e.Err.Error()
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
}
