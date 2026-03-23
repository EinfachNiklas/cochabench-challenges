# Concurrent Task Scheduler

## Task

Implement a complete concurrent task scheduler in Go that prioritizes tasks, executes them in parallel, resolves dependencies, and retries failures automatically.

The project is split into several source files that together form the scheduler:

```text
src/
├── types.go
├── priority_queue.go
├── store.go
├── worker_pool.go
├── retry.go
├── metrics.go
├── scheduler.go
└── go.mod
```

`types.go` provides the type definitions and should not be changed. The other files contain stubs that must be implemented. Additional helper files may be added if needed by the solution.

Implement the following scheduler components and behaviors:

### PriorityQueue

Implement a heap-based priority queue for tasks.

- Higher-priority tasks must be processed first
- Priority order is `PriorityCritical (3) > PriorityHigh (2) > PriorityMedium (1) > PriorityLow (0)`
- For equal priority, preserve FIFO order using `CreatedAt`
- The queue itself does not need to be thread-safe

Expected functions:

- `NewPriorityQueue()`
- `Push(task)`
- `Pop()`
- `Peek()`
- `Len()`
- `Remove(taskID)`
- `Update(taskID, newPriority)`

### TaskStore

Implement a thread-safe in-memory task store.

- Support concurrent reads and writes safely
- Store task copies instead of the original pointers
- Support filtering by status, priority, and handler

Expected functions:

- `NewTaskStore()`
- `Add(task)`
- `Get(id)`
- `Update(task)`
- `Delete(id)`
- `List(filter)`
- `Count()`

### WorkerPool

Implement a worker pool for parallel task execution.

- Support a configurable number of workers
- Each worker must consume tasks from an internal channel
- Set `Task.Status`, `Task.StartedAt`, and `Task.CompletedAt` correctly
- Support graceful shutdown so running tasks can finish

Lifecycle:

```text
NewWorkerPool(size) -> Start(ctx) -> Submit(task, handler) -> Shutdown(ctx)
```

Expected functions:

- `NewWorkerPool(size)`
- `Start(ctx)`
- `Submit(task, handler)`
- `Shutdown(ctx)`
- `ActiveWorkers()`
- `ProcessedCount()`

### Retry logic

Implement exponential backoff with a configurable retry policy.

Backoff formula:

```text
delay = min(BaseDelay * Multiplier^attempt, MaxDelay)
```

Error handling rules:

- `*RetryableError` means retry
- `*NonRetryableError` means do not retry
- Other non-nil errors should retry by default
- `nil` means no retry is needed

Expected functions:

- `CalculateDelay(attempt, policy)`
- `IsRetryable(err)`
- `WrapRetryable(err)`
- `WrapNonRetryable(err)`
- `NewRetryHandler(policy, handler)`

### MetricsCollector

Implement thread-safe task execution metrics.

- Track submissions, completions, failures, and retries
- Compute a running average execution time
- Support concurrent access safely

Running-average formula:

```text
newAvg = oldAvg + (duration - oldAvg) / totalCompleted
```

Expected functions:

- `NewMetricsCollector()`
- `RecordSubmission(priority)`
- `RecordStart(priority)`
- `RecordCompletion(priority, duration)`
- `RecordFailure(priority)`
- `RecordRetry(priority)`
- `Snapshot()`

### Scheduler

Implement the central scheduler that coordinates all components.

Provided code:

- `NewScheduler()` already contains config validation
- The `Scheduler` struct definition is already provided

Behavior to complete:

- Initialize components in `NewScheduler()`
- Register handlers
- Start the scheduler and dispatch loop
- Validate and submit tasks
- Cancel tasks
- Check task dependencies
- Detect circular dependencies
- Support graceful shutdown

Dispatch loop behavior:

1. Read tasks from the priority queue
2. Check whether all dependencies are completed
3. Requeue tasks with unmet dependencies
4. Wrap handlers with retry logic when configured
5. Submit runnable tasks to the worker pool

Dependency rules:

- Tasks may depend on other tasks through `Task.Dependencies`
- A task may run only when all dependencies are in `StatusCompleted`
- Circular dependencies must be detected during `Submit()`

## Context

Task schedulers are core components in distributed systems, CI/CD pipelines, and job queues. This challenge combines several advanced Go topics:

- Concurrency with goroutines, channels, wait groups, and atomic operations
- Heap-based data structures
- Worker-pool and backoff patterns
- Thread safety with mutexes and concurrent access
- Context-driven shutdown, timeout, and cancellation
- Error wrapping and retry classification
- Dependency resolution and cycle detection in task graphs

The challenge is intended to exercise coordination between multiple subsystems rather than a single isolated algorithm.

## Dependencies

- Go
- Use the Go standard library unless the provided module setup requires otherwise

Typical local commands:

```bash
go test ./...
```

## Constraints

- Do not change the provided public API
- Do not modify the tests
- Do not change the package name or provided type definitions
- Preserve the intended concurrency and dependency semantics
- Keep thread-safe components safe under concurrent access
- `types.go` is provided and should remain unchanged

Implementation guidance:

- Use `container/heap` for the priority queue
- Use `sync.RWMutex` where read-heavy synchronization is useful
- Respect graceful shutdown behavior in worker and scheduler coordination
- Preserve retry behavior defined by the provided error wrapper types and retry policy

## Edge Cases

- Invalid scheduler configuration
- Empty or duplicate task IDs
- Missing dependencies
- Circular dependencies
- Equal task priorities requiring FIFO ordering
- Task cancellation before execution
- Shutdown while tasks are still running
- Retryable vs. non-retryable failures
- Concurrent submissions and task execution
- Metrics snapshots during active processing
